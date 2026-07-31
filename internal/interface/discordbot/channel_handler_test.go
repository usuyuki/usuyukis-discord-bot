package discordbot

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channelcreate"
	channelcreateUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/channelcreate"
)

// fakeChannelCreator はテスト用のChannelCreatorフェイク実装
type fakeChannelCreator struct {
	gotGuildID  string
	gotCreation channelcreate.ChannelCreation
	channelID   string
	err         error
}

func (f *fakeChannelCreator) CreateChannel(ctx context.Context, guildID string, creation channelcreate.ChannelCreation) (string, error) {
	f.gotGuildID = guildID
	f.gotCreation = creation
	if f.err != nil {
		return "", f.err
	}
	return f.channelID, nil
}

func TestChannelHandler_HandleMessage(t *testing.T) {
	tests := []struct {
		name          string
		msg           IncomingMessage
		creator       *fakeChannelCreator
		wantSentSub   string
		wantCalled    bool
		wantGuildID   string
		wantPrivate   bool
		wantCreatorID string
		wantMemberIDs []string
	}{
		{
			name:          "正常系: createコマンドで公開チャンネルが作成される",
			msg:           IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "u1", MentionsBotID: true, BotID: "bot", Content: "<@bot> channel create general-2"},
			creator:       &fakeChannelCreator{channelID: "newch"},
			wantSentSub:   "公開チャンネルを作成しました: <#newch>",
			wantGuildID:   "g1",
			wantPrivate:   false,
			wantCreatorID: "u1",
		},
		{
			name:          "正常系: create-privateコマンドでメンションしたメンバーのみ閲覧可能なチャンネルが作成される",
			msg:           IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "u1", MentionsBotID: true, BotID: "bot", Content: "<@bot> channel create-private secret <@u2> <@u3>"},
			creator:       &fakeChannelCreator{channelID: "newch"},
			wantSentSub:   "プライベートチャンネルを作成しました: <#newch>",
			wantGuildID:   "g1",
			wantPrivate:   true,
			wantCreatorID: "u1",
			wantMemberIDs: []string{"u2", "u3"},
		},
		{
			name:          "正常系: create-privateコマンドはメンバー指定なしでも作成者のみ閲覧可能なチャンネルとして作成できる",
			msg:           IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "u1", MentionsBotID: true, BotID: "bot", Content: "<@bot> channel create-private secret"},
			creator:       &fakeChannelCreator{channelID: "newch"},
			wantSentSub:   "プライベートチャンネルを作成しました: <#newch>",
			wantGuildID:   "g1",
			wantPrivate:   true,
			wantCreatorID: "u1",
		},
		{
			name:        "異常系: createコマンドでチャンネル名未指定なら使い方を案内する",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "u1", MentionsBotID: true, BotID: "bot", Content: "<@bot> channel create"},
			creator:     &fakeChannelCreator{},
			wantSentSub: "使い方",
		},
		{
			name:        "異常系: ギルド外（DM等）からのコマンドは拒否される",
			msg:         IncomingMessage{GuildID: "", ChannelID: "c1", AuthorID: "u1", MentionsBotID: true, BotID: "bot", Content: "<@bot> channel create general"},
			creator:     &fakeChannelCreator{},
			wantSentSub: "サーバー内でのみ",
		},
		{
			name:        "異常系: ChannelCreatorがエラーを返すと失敗メッセージを返す",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "u1", MentionsBotID: true, BotID: "bot", Content: "<@bot> channel create general"},
			creator:     &fakeChannelCreator{err: errors.New("boom")},
			wantSentSub: "チャンネル作成に失敗しました",
		},
		{
			name:       "正常系: channelで始まらないコマンドは無視する",
			msg:        IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "u1", MentionsBotID: true, BotID: "bot", Content: "<@bot> keyword list"},
			creator:    &fakeChannelCreator{},
			wantCalled: false,
		},
		{
			name:       "正常系: Botへのメンションがなければ反応しない",
			msg:        IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "u1", MentionsBotID: false, BotID: "bot", Content: "channel create general"},
			creator:    &fakeChannelCreator{},
			wantCalled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := channelcreateUC.New(tt.creator)
			sender := &fakeMessageSender{}
			h := NewChannelHandler(uc, sender)

			if err := h.HandleMessage(context.Background(), tt.msg); err != nil {
				t.Fatalf("HandleMessage() unexpected error = %v", err)
			}

			wantCalled := tt.wantCalled || tt.wantSentSub != ""
			if sender.called != wantCalled {
				t.Fatalf("HandleMessage() sender.called = %v, want %v", sender.called, wantCalled)
			}
			if tt.wantSentSub != "" && !strings.Contains(sender.sentContent, tt.wantSentSub) {
				t.Errorf("HandleMessage() content = %q, want substring %q", sender.sentContent, tt.wantSentSub)
			}
			if tt.wantCreatorID == "" {
				return
			}
			if tt.creator.gotGuildID != tt.wantGuildID {
				t.Errorf("CreateChannel() guildID = %q, want %q", tt.creator.gotGuildID, tt.wantGuildID)
			}
			if tt.creator.gotCreation.Private != tt.wantPrivate {
				t.Errorf("CreateChannel() creation.Private = %v, want %v", tt.creator.gotCreation.Private, tt.wantPrivate)
			}
			if tt.creator.gotCreation.CreatorID != tt.wantCreatorID {
				t.Errorf("CreateChannel() creation.CreatorID = %q, want %q", tt.creator.gotCreation.CreatorID, tt.wantCreatorID)
			}
			if !reflect.DeepEqual(tt.creator.gotCreation.MemberIDs, tt.wantMemberIDs) {
				t.Errorf("CreateChannel() creation.MemberIDs = %v, want %v", tt.creator.gotCreation.MemberIDs, tt.wantMemberIDs)
			}
		})
	}
}

func TestParseChannelCommand(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *channelCommand
	}{
		{
			name:    "正常系: createコマンドを解析できる",
			content: "<@bot> channel create general-2",
			want:    &channelCommand{Sub: "create", Name: "general-2"},
		},
		{
			name:    "正常系: create-privateコマンドはメンバーのメンションを解析できる",
			content: "<@bot> channel create-private secret <@u2> <@u3>",
			want:    &channelCommand{Sub: "create-private", Name: "secret", MemberIDs: []string{"u2", "u3"}},
		},
		{
			name:    "正常系: create-privateコマンドはメンション無しでも解析できる",
			content: "<@bot> channel create-private secret",
			want:    &channelCommand{Sub: "create-private", Name: "secret"},
		},
		{
			name:    "異常系: channelで始まらない場合はnilを返す",
			content: "<@bot> keyword list",
			want:    nil,
		},
		{
			name:    "異常系: サブコマンドがない場合はnilを返す",
			content: "<@bot> channel",
			want:    nil,
		},
		{
			name:    "異常系: 未知のサブコマンドはnilを返す",
			content: "<@bot> channel delete general",
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseChannelCommand(tt.content, "bot")
			if (got == nil) != (tt.want == nil) {
				t.Fatalf("parseChannelCommand(%q) = %+v, want %+v", tt.content, got, tt.want)
			}
			if got == nil {
				return
			}
			if got.Sub != tt.want.Sub || got.Name != tt.want.Name || !reflect.DeepEqual(got.MemberIDs, tt.want.MemberIDs) {
				t.Errorf("parseChannelCommand(%q) = %+v, want %+v", tt.content, got, tt.want)
			}
		})
	}
}

func TestExtractMentionedUserIDs(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		want   []string
	}{
		{
			name:   "正常系: 構造化メンションのみをIDへ変換する",
			fields: []string{"<@u1>", "not-a-mention", "<@u2>"},
			want:   []string{"u1", "u2"},
		},
		{
			name:   "正常系: ニックネームメンション形式（<@!id>）も解析できる",
			fields: []string{"<@!u1>"},
			want:   []string{"u1"},
		},
		{
			name:   "異常系: メンションが1つもなければnilを返す",
			fields: []string{"secret", "channel"},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractMentionedUserIDs(tt.fields)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractMentionedUserIDs(%v) = %v, want %v", tt.fields, got, tt.want)
			}
		})
	}
}
