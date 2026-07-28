package discordbot

import (
	"context"
	"strings"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
	keywordUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/keyword"
)

type fakeKeywordRepository struct {
	items map[string][]keyword.Keyword
}

func newFakeKeywordRepository() *fakeKeywordRepository {
	return &fakeKeywordRepository{items: map[string][]keyword.Keyword{}}
}

func (f *fakeKeywordRepository) Save(ctx context.Context, k keyword.Keyword) error {
	f.items[k.GuildID] = append(f.items[k.GuildID], k)
	return nil
}

func (f *fakeKeywordRepository) Delete(ctx context.Context, guildID, word string) error {
	kept := f.items[guildID][:0]
	for _, k := range f.items[guildID] {
		if k.Word != word {
			kept = append(kept, k)
		}
	}
	f.items[guildID] = kept
	return nil
}

func (f *fakeKeywordRepository) FindByGuild(ctx context.Context, guildID string) ([]keyword.Keyword, error) {
	return f.items[guildID], nil
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
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: true, Content: "<@bot> keyword add ぬるぽ ガッ"},
			wantSentSub: "登録しました",
		},
		{
			name:        "異常系: 管理者でない場合はaddコマンドが拒否される",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: false, Content: "<@bot> keyword add ぬるぽ ガッ"},
			wantSentSub: "管理者権限が必要です",
		},
		{
			name:        "正常系: listコマンドで登録済みキーワードが一覧表示される",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: false, Content: "<@bot> keyword list"},
			preRegister: true,
			wantSentSub: "ぬるぽ",
		},
		{
			name:        "異常系: removeコマンドで単語未指定なら使い方を案内する",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true, IsAdmin: true, Content: "<@bot> keyword remove"},
			wantSentSub: "使い方",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeKeywordRepository()
			if tt.preRegister {
				k, _ := keyword.New(0, "g1", "ぬるぽ", "ガッ")
				repo.items["g1"] = append(repo.items["g1"], k)
			}
			uc := keywordUC.New(repo)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeKeywordRepository()
			k, _ := keyword.New(0, "g1", "ぬるぽ", "ガッ")
			repo.items["g1"] = append(repo.items["g1"], k)
			uc := keywordUC.New(repo)
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
			name:    "正常系: 複数メンションが含まれていてもすべて除去して解析できる",
			content: "<@bot> <@&role1> keyword add ぬるぽ ガッ",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseKeywordCommand(tt.content)
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
