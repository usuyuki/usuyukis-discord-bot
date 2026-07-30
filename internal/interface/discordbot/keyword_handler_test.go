package discordbot

import (
	"context"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
	keywordUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/keyword"
)

// fakeKeywordRandomizer はmath/rand/v2経由でRandomizer portを満たすテスト用実装。
// 乱数選択が実際に複数の応答候補を行き来することを検証するテストで使う
type fakeKeywordRandomizer struct{}

func (fakeKeywordRandomizer) Intn(n int) int {
	return rand.IntN(n)
}

// fakeKeywordRepository はテスト用のインメモリRepository実装。
// guildID -> word -> responses の形で保持し、応答の積み増し・個別削除を再現する
type fakeKeywordRepository struct {
	items map[string]map[string][]string
}

func newFakeKeywordRepository() *fakeKeywordRepository {
	return &fakeKeywordRepository{items: map[string]map[string][]string{}}
}

func (f *fakeKeywordRepository) AddResponse(ctx context.Context, guildID, word, response string) error {
	if f.items[guildID] == nil {
		f.items[guildID] = map[string][]string{}
	}
	f.items[guildID][word] = append(f.items[guildID][word], response)
	return nil
}

func (f *fakeKeywordRepository) RemoveResponse(ctx context.Context, guildID, word, response string) error {
	responses := f.items[guildID][word]
	kept := responses[:0]
	for _, r := range responses {
		if r != response {
			kept = append(kept, r)
		}
	}
	if len(kept) == 0 {
		delete(f.items[guildID], word)
		return nil
	}
	f.items[guildID][word] = kept
	return nil
}

func (f *fakeKeywordRepository) RemoveKeyword(ctx context.Context, guildID, word string) error {
	delete(f.items[guildID], word)
	return nil
}

func (f *fakeKeywordRepository) ReplaceResponses(ctx context.Context, guildID, word string, responses []string) error {
	if f.items[guildID] == nil {
		f.items[guildID] = map[string][]string{}
	}
	f.items[guildID][word] = responses
	return nil
}

func (f *fakeKeywordRepository) FindByGuild(ctx context.Context, guildID string) ([]keyword.Keyword, error) {
	var result []keyword.Keyword
	for word, responses := range f.items[guildID] {
		k, err := keyword.New(0, guildID, word, responses)
		if err != nil {
			return nil, err
		}
		result = append(result, k)
	}
	return result, nil
}

func TestKeywordHandler_HandleMessage_Command(t *testing.T) {
	tests := []struct {
		name        string
		msg         IncomingMessage
		preRegister bool
		wantSentSub string // 送信された文言に含まれるべき文字列
	}{
		{
			name:        "正常系: 管理者がaddコマンドを実行すると登録され、登録完了メッセージが返る",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: true, BotID: "bot", Content: "<@bot> keyword add ぬるぽ ガッ"},
			wantSentSub: "登録しました",
		},
		{
			name:        "異常系: 管理者でない場合はaddコマンドが拒否される",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: false, BotID: "bot", Content: "<@bot> keyword add ぬるぽ ガッ"},
			wantSentSub: "管理者権限が必要です",
		},
		{
			name:        "正常系: listコマンドで登録済みキーワードが一覧表示される",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: false, BotID: "bot", Content: "<@bot> keyword list"},
			preRegister: true,
			wantSentSub: "ぬるぽ",
		},
		{
			name:        "異常系: removeコマンドで単語未指定なら使い方を案内する",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: true, BotID: "bot", Content: "<@bot> keyword remove"},
			wantSentSub: "使い方",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeKeywordRepository()
			if tt.preRegister {
				repo.items["g1"] = map[string][]string{"ぬるぽ": {"ガッ"}}
			}
			uc := keywordUC.New(repo, fakeKeywordRandomizer{})
			sender := &fakeMessageSender{}
			h := NewKeywordHandler(uc, sender)

			if err := h.HandleMessage(context.Background(), tt.msg); err != nil {
				t.Fatalf("HandleMessage() unexpected error = %v", err)
			}
			if !sender.called {
				t.Fatal("HandleMessage() expected a reply to be sent, but none was")
			}
			if !strings.Contains(sender.sentContent, tt.wantSentSub) {
				t.Errorf("HandleMessage() content = %q, want substring %q", sender.sentContent, tt.wantSentSub)
			}
		})
	}
}

func TestKeywordHandler_HandleMessage_Command_MultipleResponses(t *testing.T) {
	repo := newFakeKeywordRepository()
	uc := keywordUC.New(repo, fakeKeywordRandomizer{})
	sender := &fakeMessageSender{}
	h := NewKeywordHandler(uc, sender)
	ctx := context.Background()

	addMsg := IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: true, BotID: "bot", Content: "<@bot> keyword add ぬるぽ ガッ"}
	if err := h.HandleMessage(ctx, addMsg); err != nil {
		t.Fatalf("HandleMessage() unexpected error = %v", err)
	}
	addMsg2 := IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: true, BotID: "bot", Content: "<@bot> keyword add ぬるぽ ｶﾞｯ"}
	if err := h.HandleMessage(ctx, addMsg2); err != nil {
		t.Fatalf("HandleMessage() unexpected error = %v", err)
	}

	if len(repo.items["g1"]["ぬるぽ"]) != 2 {
		t.Fatalf("同じキーワードへの複数回addは応答を積み増すはずが、%v件しか登録されていない", len(repo.items["g1"]["ぬるぽ"]))
	}

	removeMsg := IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: true, BotID: "bot", Content: "<@bot> keyword remove ぬるぽ ガッ"}
	if err := h.HandleMessage(ctx, removeMsg); err != nil {
		t.Fatalf("HandleMessage() unexpected error = %v", err)
	}
	if !strings.Contains(sender.sentContent, "削除しました") {
		t.Fatalf("HandleMessage() content = %q, want substring %q", sender.sentContent, "削除しました")
	}
	if got := repo.items["g1"]["ぬるぽ"]; len(got) != 1 || got[0] != "ｶﾞｯ" {
		t.Fatalf("remove後の応答一覧 = %v, want [ｶﾞｯ] のみ残っている", got)
	}
}

func TestKeywordHandler_HandleMessage_UsageMessage_EmbedsBotIDDynamically(t *testing.T) {
	repo := newFakeKeywordRepository()
	uc := keywordUC.New(repo, fakeKeywordRandomizer{})
	sender := &fakeMessageSender{}
	h := NewKeywordHandler(uc, sender)

	msg := IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, BotID: "999", IsAdmin: true, Content: "<@999> keyword add"}
	if err := h.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage() unexpected error = %v", err)
	}
	want := "<@999>"
	if !strings.Contains(sender.sentContent, want) {
		t.Errorf("HandleMessage() content = %q, want to contain %q (使い方メッセージのメンション表記はmsg.BotIDから動的に組み立てられるべき)", sender.sentContent, want)
	}
}

func TestKeywordHandler_HandleMessage_NoBotMention_FallsBackToAutoReply(t *testing.T) {
	repo := newFakeKeywordRepository()
	uc := keywordUC.New(repo, fakeKeywordRandomizer{})
	sender := &fakeMessageSender{}
	h := NewKeywordHandler(uc, sender)

	// 構造化メンションを含まない発言は、たとえ"keyword add"という文字列を含んでいても
	// コマンドとして解釈されず、キーワード自動応答としてのマッチのみ試みられる（未登録なので無応答）
	msg := IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false, BotID: "bot", IsAdmin: true, Content: "keyword add ぬるぽ ガッ"}
	if err := h.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage() unexpected error = %v", err)
	}
	if sender.called {
		t.Errorf("HandleMessage() sent a reply %q, want no reply", sender.sentContent)
	}
	if len(repo.items["g1"]["ぬるぽ"]) != 0 {
		t.Errorf("keyword addがコマンドとして誤って処理されている: repo.items[g1][ぬるぽ] = %v", repo.items["g1"]["ぬるぽ"])
	}
}

func TestKeywordHandler_HandleMessage_AutoReply(t *testing.T) {
	tests := []struct {
		name       string
		msg        IncomingMessage
		wantCalled bool
		wantReply  string
	}{
		{
			name:       "正常系: 通常メッセージが登録済みキーワードに一致すれば応答する",
			msg:        IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false, Content: "さっきぬるぽ食らった"},
			wantCalled: true,
			wantReply:  "ガッ",
		},
		{
			name:       "異常系: 一致しなければ何も送信しない",
			msg:        IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false, Content: "今日はいい天気"},
			wantCalled: false,
		},
		{
			name:       "正常系: @Rakuroはコマンド起動用の予約語ではなく、通常のキーワードとして自動応答にマッチする",
			msg:        IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false, Content: "@Rakuro 今何時？"},
			wantCalled: true,
			wantReply:  "呼んだ？",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeKeywordRepository()
			repo.items["g1"] = map[string][]string{"ぬるぽ": {"ガッ"}, "@Rakuro": {"呼んだ？"}}
			uc := keywordUC.New(repo, fakeKeywordRandomizer{})
			sender := &fakeMessageSender{}
			h := NewKeywordHandler(uc, sender)

			if err := h.HandleMessage(context.Background(), tt.msg); err != nil {
				t.Fatalf("HandleMessage() unexpected error = %v", err)
			}
			if sender.called != tt.wantCalled {
				t.Fatalf("HandleMessage() called = %v, want %v", sender.called, tt.wantCalled)
			}
			if tt.wantCalled && sender.sentContent != tt.wantReply {
				t.Errorf("HandleMessage() content = %q, want %q", sender.sentContent, tt.wantReply)
			}
		})
	}
}

func TestKeywordHandler_HandleMessage_AutoReply_RandomlyPicksResponse(t *testing.T) {
	repo := newFakeKeywordRepository()
	repo.items["g1"] = map[string][]string{"ぬるぽ": {"ガッ", "ｶﾞｯ"}}
	uc := keywordUC.New(repo, fakeKeywordRandomizer{})
	sender := &fakeMessageSender{}
	h := NewKeywordHandler(uc, sender)

	seen := map[string]bool{}
	for range 50 {
		msg := IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false, Content: "さっきぬるぽ食らった"}
		if err := h.HandleMessage(context.Background(), msg); err != nil {
			t.Fatalf("HandleMessage() unexpected error = %v", err)
		}
		if sender.sentContent != "ガッ" && sender.sentContent != "ｶﾞｯ" {
			t.Fatalf("HandleMessage() sentContent = %q, want ガッ or ｶﾞｯ", sender.sentContent)
		}
		seen[sender.sentContent] = true
	}
	if len(seen) < 2 {
		t.Fatalf("50回試行して応答が%vのみ、複数応答からランダムに選ばれていない可能性がある", seen)
	}
}

func TestKeywordHandler_HandleMessage_AutoReply_ExpandsNowPlaceholder(t *testing.T) {
	repo := newFakeKeywordRepository()
	repo.items["g1"] = map[string][]string{"今何時": {"今は{$now}だよ"}}
	uc := keywordUC.New(repo, fakeKeywordRandomizer{})
	sender := &fakeMessageSender{}
	h := NewKeywordHandler(uc, sender)
	h.now = func() time.Time { return time.Date(2026, 7, 28, 14, 32, 10, 0, time.UTC) }

	msg := IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false, Content: "今何時"}
	if err := h.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage() unexpected error = %v", err)
	}
	want := "今は2026-07-28 23:32:10だよ"
	if sender.sentContent != want {
		t.Fatalf("HandleMessage() content = %q, want %q", sender.sentContent, want)
	}
}

func TestParseKeywordCommand(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *keywordCommand
	}{
		{
			name:    "正常系: メンション1つに続くaddコマンドを解析できる",
			content: "<@bot> keyword add ぬるぽ ガッ",
			want:    &keywordCommand{Sub: "add", Word: "ぬるぽ", Response: "ガッ"},
		},
		{
			name:    "正常系: Botメンションが先頭以外にもあればすべて除去して解析できる",
			content: "<@bot> keyword <@bot> add ぬるぽ ガッ",
			want:    &keywordCommand{Sub: "add", Word: "ぬるぽ", Response: "ガッ"},
		},
		{
			name:    "正常系: メンションが末尾にあってもkeywordコマンドとして解析できる",
			content: "keyword list <@bot>",
			want:    &keywordCommand{Sub: "list"},
		},
		{
			name:    "正常系: 全角スペース区切りでもstrings.FieldsがUnicode空白として認識し解析できる",
			content: "<@bot>　keyword　add",
			want:    &keywordCommand{Sub: "add"},
		},
		{
			name:    "異常系: keywordで始まらない場合はnilを返す",
			content: "<@bot> hello world",
			want:    nil,
		},
		{
			name:    "異常系: メンションのみでサブコマンドがない場合はnilを返す",
			content: "<@bot> keyword",
			want:    nil,
		},
		{
			name:    "異常系: 未知のサブコマンドはnilを返す",
			content: "<@bot> keyword rename ぬるぽ",
			want:    nil,
		},
		{
			name:    "異常系: 本文が空の場合はnilを返す",
			content: "",
			want:    nil,
		},
		{
			name:    "異常系: テキストの@メンション風表記は構造化メンションでないため除去されず、keyword以外の先頭語としてnilを返す",
			content: "@bot keyword add ぬるぽ ガッ",
			want:    nil,
		},
		{
			name:    "正常系: removeコマンドはキーワードと応答の2引数を解析できる",
			content: "<@bot> keyword remove ぬるぽ ガッ",
			want:    &keywordCommand{Sub: "remove", Word: "ぬるぽ", Response: "ガッ"},
		},
		{
			name:    "正常系: Botメンションと同じ形式でもBotID以外のメンション風文字列はキーワード引数として保持される",
			content: "<@bot> keyword add <@notanid> ガッ",
			want:    &keywordCommand{Sub: "add", Word: "<@notanid>", Response: "ガッ"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseKeywordCommand(tt.content, "bot")
			if (got == nil) != (tt.want == nil) {
				t.Fatalf("parseKeywordCommand(%q) = %+v, want %+v", tt.content, got, tt.want)
			}
			if got == nil {
				return
			}
			if *got != *tt.want {
				t.Errorf("parseKeywordCommand(%q) = %+v, want %+v", tt.content, got, tt.want)
			}
		})
	}
}
