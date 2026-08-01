package discordbot

import (
	"context"
	"errors"
	"testing"
)

// fakeChannelProposeUseCase はテスト用のUseCaseフェイク実装。呼び出されたかと引数を記録する
type fakeChannelProposeUseCase struct {
	called        bool
	gotGuildID    string
	gotChannelID  string
	gotProposerID string
	gotName       string
	err           error
}

func (f *fakeChannelProposeUseCase) Propose(ctx context.Context, guildID, channelID, proposerID, name string) error {
	f.called = true
	f.gotGuildID = guildID
	f.gotChannelID = channelID
	f.gotProposerID = proposerID
	f.gotName = name
	return f.err
}

func TestChannelHandler_HandleMessage(t *testing.T) {
	const botID = "bot1"

	t.Run("正常系: channel createコマンドでUseCaseへ委譲する（成功時は提案メッセージ自体が通知を兼ねるため追加送信しない）", func(t *testing.T) {
		uc := &fakeChannelProposeUseCase{}
		sender := &fakeMessageSender{}
		h := NewChannelHandler(uc, sender)

		msg := IncomingMessage{
			GuildID:       "g1",
			ChannelID:     "c1",
			AuthorID:      "user1",
			Content:       "<@bot1> channel create general-chat",
			MentionsBotID: true,
			BotID:         botID,
		}
		if err := h.HandleMessage(context.Background(), msg); err != nil {
			t.Fatalf("HandleMessage() unexpected error = %v", err)
		}
		if !uc.called {
			t.Fatal("HandleMessage() should call UseCase for channel create command")
		}
		if uc.gotGuildID != "g1" || uc.gotChannelID != "c1" || uc.gotProposerID != "user1" || uc.gotName != "general-chat" {
			t.Errorf("UseCase received guildID=%q channelID=%q proposerID=%q name=%q, want g1/c1/user1/general-chat", uc.gotGuildID, uc.gotChannelID, uc.gotProposerID, uc.gotName)
		}
		if sender.called {
			t.Error("HandleMessage() should not send an extra message on success")
		}
	})

	t.Run("異常系: channel createコマンドの名前が未指定だと使い方を案内しUseCaseは呼ばれない", func(t *testing.T) {
		uc := &fakeChannelProposeUseCase{}
		sender := &fakeMessageSender{}
		h := NewChannelHandler(uc, sender)

		msg := IncomingMessage{
			GuildID:       "g1",
			ChannelID:     "c1",
			Content:       "<@bot1> channel create",
			MentionsBotID: true,
			BotID:         botID,
		}
		if err := h.HandleMessage(context.Background(), msg); err != nil {
			t.Fatalf("HandleMessage() unexpected error = %v", err)
		}
		if uc.called {
			t.Error("HandleMessage() should not call UseCase when name is missing")
		}
		if !sender.called {
			t.Fatal("expected a usage message to be sent")
		}
	})

	t.Run("異常系: UseCaseがエラーを返すとエラー内容を通知しエラーは返さない（ユーザー起因の入力ミスのため）", func(t *testing.T) {
		uc := &fakeChannelProposeUseCase{err: errors.New("channel: name must contain only lowercase letters, numbers, hyphens, underscores, or Japanese characters")}
		sender := &fakeMessageSender{}
		h := NewChannelHandler(uc, sender)

		msg := IncomingMessage{
			GuildID:       "g1",
			ChannelID:     "c1",
			Content:       "<@bot1> channel create Invalid!",
			MentionsBotID: true,
			BotID:         botID,
		}
		if err := h.HandleMessage(context.Background(), msg); err != nil {
			t.Fatalf("HandleMessage() unexpected error = %v", err)
		}
		if !sender.called {
			t.Fatal("expected an error message to be sent")
		}
	})

	t.Run("異常系: Botへのメンションがなければ何もしない", func(t *testing.T) {
		uc := &fakeChannelProposeUseCase{}
		sender := &fakeMessageSender{}
		h := NewChannelHandler(uc, sender)

		msg := IncomingMessage{
			GuildID:       "g1",
			ChannelID:     "c1",
			Content:       "channel create general-chat",
			MentionsBotID: false,
			BotID:         botID,
		}
		if err := h.HandleMessage(context.Background(), msg); err != nil {
			t.Fatalf("HandleMessage() unexpected error = %v", err)
		}
		if uc.called {
			t.Error("HandleMessage() should not call UseCase without a bot mention")
		}
	})

	t.Run("正常系: 構造化メンションでなく地の文の@botName表記でもchannel createコマンドを認識する（コピペ救済）", func(t *testing.T) {
		uc := &fakeChannelProposeUseCase{}
		sender := &fakeMessageSender{}
		h := NewChannelHandler(uc, sender)

		msg := IncomingMessage{
			GuildID:       "g1",
			ChannelID:     "c1",
			AuthorID:      "user1",
			Content:       "@usuyuki channel create general-chat",
			MentionsBotID: true,
			BotID:         botID,
			BotName:       "usuyuki",
		}
		if err := h.HandleMessage(context.Background(), msg); err != nil {
			t.Fatalf("HandleMessage() unexpected error = %v", err)
		}
		if !uc.called {
			t.Fatal("HandleMessage() should call UseCase for channel create command via botName fallback")
		}
		if uc.gotName != "general-chat" {
			t.Errorf("UseCase received name=%q, want general-chat", uc.gotName)
		}
	})

	t.Run("異常系: channel以外のコマンドには反応しない", func(t *testing.T) {
		uc := &fakeChannelProposeUseCase{}
		sender := &fakeMessageSender{}
		h := NewChannelHandler(uc, sender)

		msg := IncomingMessage{
			GuildID:       "g1",
			ChannelID:     "c1",
			Content:       "<@bot1> help",
			MentionsBotID: true,
			BotID:         botID,
		}
		if err := h.HandleMessage(context.Background(), msg); err != nil {
			t.Fatalf("HandleMessage() unexpected error = %v", err)
		}
		if uc.called {
			t.Error("HandleMessage() should not call UseCase for unrelated commands")
		}
		if sender.called {
			t.Error("HandleMessage() should not send any message for unrelated commands")
		}
	})
}
