package discordbot

import (
	"context"
	"errors"
	"strings"
	"testing"

	channelUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/channel"
)

// fakeChannelCreator はテスト用のCreatorフェイク実装。CreatePrivateChannelに渡された引数を記録する
type fakeChannelCreator struct {
	gotGuildID   string
	gotName      string
	gotCreatorID string
	channelID    string
	err          error
}

func (f *fakeChannelCreator) CreatePrivateChannel(ctx context.Context, guildID, name, creatorUserID string) (string, error) {
	f.gotGuildID = guildID
	f.gotName = name
	f.gotCreatorID = creatorUserID
	return f.channelID, f.err
}

func TestChannelHandler_HandleMessage(t *testing.T) {
	tests := []struct {
		name        string
		msg         IncomingMessage
		creator     *fakeChannelCreator
		wantCalled  bool
		wantSentSub string
	}{
		{
			name:        "正常系: 管理者でない一般ユーザーでもcreateコマンドでチャンネルを作成できる",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "user1", MentionsBotID: true, IsAdmin: false, BotID: "bot", Content: "<@bot> channel create 雑談"},
			creator:     &fakeChannelCreator{channelID: "newch"},
			wantCalled:  true,
			wantSentSub: "<#newch>",
		},
		{
			name:        "異常系: チャンネル名を省略すると使い方を案内しCreatorは呼ばれない",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "user1", MentionsBotID: true, BotID: "bot", Content: "<@bot> channel create"},
			creator:     &fakeChannelCreator{},
			wantCalled:  true,
			wantSentSub: "使い方",
		},
		{
			name:       "異常系: 未知のサブコマンドには反応しない",
			msg:        IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "user1", MentionsBotID: true, BotID: "bot", Content: "<@bot> channel rename foo"},
			creator:    &fakeChannelCreator{},
			wantCalled: false,
		},
		{
			name:       "異常系: channelで始まらない本文には反応しない",
			msg:        IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "user1", MentionsBotID: true, BotID: "bot", Content: "<@bot> keyword list"},
			creator:    &fakeChannelCreator{},
			wantCalled: false,
		},
		{
			name:       "異常系: メンションがないと反応しない",
			msg:        IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "user1", MentionsBotID: false, BotID: "bot", Content: "channel create 雑談"},
			creator:    &fakeChannelCreator{},
			wantCalled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := channelUC.New(tt.creator)
			sender := &fakeMessageSender{}
			h := NewChannelHandler(uc, sender)

			if err := h.HandleMessage(context.Background(), tt.msg); err != nil {
				t.Fatalf("HandleMessage() unexpected error = %v", err)
			}
			if sender.called != tt.wantCalled {
				t.Fatalf("HandleMessage() called = %v, want %v", sender.called, tt.wantCalled)
			}
			if tt.wantCalled && !strings.Contains(sender.sentContent, tt.wantSentSub) {
				t.Errorf("HandleMessage() content = %q, want substring %q", sender.sentContent, tt.wantSentSub)
			}
		})
	}
}

func TestChannelHandler_HandleMessage_PassesGuildIDAndAuthorIDToCreator(t *testing.T) {
	creator := &fakeChannelCreator{channelID: "newch"}
	uc := channelUC.New(creator)
	sender := &fakeMessageSender{}
	h := NewChannelHandler(uc, sender)

	msg := IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "user1", MentionsBotID: true, BotID: "bot", Content: "<@bot> channel create 雑談"}
	if err := h.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage() unexpected error = %v", err)
	}
	if creator.gotGuildID != "g1" || creator.gotCreatorID != "user1" || creator.gotName != "雑談" {
		t.Errorf("Creator received guildID=%q creatorID=%q name=%q, want g1/user1/雑談", creator.gotGuildID, creator.gotCreatorID, creator.gotName)
	}
}

func TestChannelHandler_HandleMessage_CreatorError_IsPropagated(t *testing.T) {
	creator := &fakeChannelCreator{err: errors.New("discord api boom")}
	uc := channelUC.New(creator)
	sender := &fakeMessageSender{}
	h := NewChannelHandler(uc, sender)

	msg := IncomingMessage{GuildID: "g1", ChannelID: "c1", AuthorID: "user1", MentionsBotID: true, BotID: "bot", Content: "<@bot> channel create 雑談"}
	err := h.HandleMessage(context.Background(), msg)
	if err == nil {
		t.Fatal("HandleMessage() expected error from Creator to be propagated, got nil")
	}
	if sender.called {
		t.Errorf("HandleMessage() should not send a reply when Creator fails, but sent %q", sender.sentContent)
	}
}

func TestParseChannelCommand(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *channelCommand
	}{
		{
			name:    "正常系: メンションに続くcreateコマンドを解析できる",
			content: "<@bot> channel create 雑談",
			want:    &channelCommand{Sub: "create", Name: "雑談"},
		},
		{
			name:    "正常系: チャンネル名省略時はNameが空文字になる",
			content: "<@bot> channel create",
			want:    &channelCommand{Sub: "create", Name: ""},
		},
		{
			name:    "異常系: channelで始まらない場合はnilを返す",
			content: "<@bot> hello world",
			want:    nil,
		},
		{
			name:    "異常系: サブコマンドがない場合はnilを返す",
			content: "<@bot> channel",
			want:    nil,
		},
		{
			name:    "異常系: 未知のサブコマンドはnilを返す",
			content: "<@bot> channel rename foo",
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
			got := parseChannelCommand(tt.content, "bot")
			if (got == nil) != (tt.want == nil) {
				t.Fatalf("parseChannelCommand(%q) = %+v, want %+v", tt.content, got, tt.want)
			}
			if got == nil {
				return
			}
			if *got != *tt.want {
				t.Errorf("parseChannelCommand(%q) = %+v, want %+v", tt.content, got, tt.want)
			}
		})
	}
}
