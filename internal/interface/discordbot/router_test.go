package discordbot

import (
	"context"
	"errors"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/emoji"
)

type recordingMessageHandler struct {
	received []IncomingMessage
	err      error
}

func (h *recordingMessageHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	h.received = append(h.received, msg)
	return h.err
}

type recordingEmojiHandler struct {
	received []IncomingEmojiUpdate
}

func (h *recordingEmojiHandler) HandleEmojiUpdate(ctx context.Context, ev IncomingEmojiUpdate) error {
	h.received = append(h.received, ev)
	return nil
}

type recordingReactionAddHandler struct {
	received []IncomingReactionAdd
}

func (h *recordingReactionAddHandler) HandleReactionAdd(ctx context.Context, ev IncomingReactionAdd) error {
	h.received = append(h.received, ev)
	return nil
}

func TestRouter_DispatchMessage(t *testing.T) {
	t.Run("正常系: 登録済み全ハンドラにメッセージが配送される", func(t *testing.T) {
		r := NewRouter()
		h1 := &recordingMessageHandler{}
		h2 := &recordingMessageHandler{}
		r.RegisterMessageHandler(h1)
		r.RegisterMessageHandler(h2)

		msg := IncomingMessage{GuildID: "g1", Content: "hello"}
		r.DispatchMessage(context.Background(), msg)

		if len(h1.received) != 1 || h1.received[0] != msg {
			t.Errorf("h1 did not receive expected message: %v", h1.received)
		}
		if len(h2.received) != 1 || h2.received[0] != msg {
			t.Errorf("h2 did not receive expected message: %v", h2.received)
		}
	})

	t.Run("異常系: 1つのハンドラがエラーを返しても他のハンドラは実行される", func(t *testing.T) {
		r := NewRouter()
		failing := &recordingMessageHandler{err: errors.New("boom")}
		ok := &recordingMessageHandler{}
		r.RegisterMessageHandler(failing)
		r.RegisterMessageHandler(ok)

		r.DispatchMessage(context.Background(), IncomingMessage{Content: "x"})

		if len(ok.received) != 1 {
			t.Errorf("ok handler should still receive the message even if a preceding handler fails, got %v", ok.received)
		}
	})

	t.Run("正常系: dev mode未設定時はどのチャンネルの投稿もハンドラに配送される", func(t *testing.T) {
		r := NewRouter()
		h := &recordingMessageHandler{}
		r.RegisterMessageHandler(h)

		r.DispatchMessage(context.Background(), IncomingMessage{ChannelID: "any-channel"})

		if len(h.received) != 1 {
			t.Errorf("handler should receive the message when dev mode is disabled, got %v", h.received)
		}
	})

	t.Run("正常系: dev mode設定時は指定チャンネルの投稿がハンドラに配送される", func(t *testing.T) {
		r := NewRouter()
		r.SetDevChannelID("dev-channel")
		h := &recordingMessageHandler{}
		r.RegisterMessageHandler(h)

		r.DispatchMessage(context.Background(), IncomingMessage{ChannelID: "dev-channel"})

		if len(h.received) != 1 {
			t.Errorf("handler should receive the message from the dev channel, got %v", h.received)
		}
	})

	t.Run("異常系: dev mode設定時は指定チャンネル以外の投稿はハンドラに配送されない", func(t *testing.T) {
		r := NewRouter()
		r.SetDevChannelID("dev-channel")
		h := &recordingMessageHandler{}
		r.RegisterMessageHandler(h)

		r.DispatchMessage(context.Background(), IncomingMessage{ChannelID: "other-channel"})

		if len(h.received) != 0 {
			t.Errorf("handler should not receive the message from a non-dev channel, got %v", h.received)
		}
	})
}

func TestRouter_DispatchEmojiUpdate(t *testing.T) {
	newEmoji := func(t *testing.T) emoji.Emoji {
		e, err := emoji.New("new", "999", false)
		if err != nil {
			t.Fatalf("emoji.New() unexpected error = %v", err)
		}
		return e
	}

	t.Run("正常系: 登録済み全ハンドラに絵文字更新イベントが配送される", func(t *testing.T) {
		r := NewRouter()
		h := &recordingEmojiHandler{}
		r.RegisterEmojiUpdateHandler(h)

		ev := IncomingEmojiUpdate{GuildID: "g1", AddedEmojis: []emoji.Emoji{newEmoji(t)}}
		r.DispatchEmojiUpdate(context.Background(), ev)

		if len(h.received) != 1 || h.received[0].GuildID != "g1" {
			t.Errorf("handler did not receive expected event: %v", h.received)
		}
	})

	t.Run("異常系: dev mode設定時は絵文字更新イベントはチャンネル概念を持たないため配送されない", func(t *testing.T) {
		r := NewRouter()
		r.SetDevChannelID("dev-channel")
		h := &recordingEmojiHandler{}
		r.RegisterEmojiUpdateHandler(h)

		ev := IncomingEmojiUpdate{GuildID: "g1", AddedEmojis: []emoji.Emoji{newEmoji(t)}}
		r.DispatchEmojiUpdate(context.Background(), ev)

		if len(h.received) != 0 {
			t.Errorf("handler should not receive emoji update events while dev mode is enabled, got %v", h.received)
		}
	})
}

func TestRouter_DispatchReactionAdd(t *testing.T) {
	t.Run("正常系: 登録済み全ハンドラにリアクション追加イベントが配送される", func(t *testing.T) {
		r := NewRouter()
		h := &recordingReactionAddHandler{}
		r.RegisterReactionAddHandler(h)

		ev := IncomingReactionAdd{ChannelID: "c1", MessageID: "msg1"}
		r.DispatchReactionAdd(context.Background(), ev)

		if len(h.received) != 1 || h.received[0] != ev {
			t.Errorf("handler did not receive expected event: %v", h.received)
		}
	})

	t.Run("異常系: dev mode設定時は指定チャンネル以外のリアクションは配送されない", func(t *testing.T) {
		r := NewRouter()
		r.SetDevChannelID("dev-channel")
		h := &recordingReactionAddHandler{}
		r.RegisterReactionAddHandler(h)

		r.DispatchReactionAdd(context.Background(), IncomingReactionAdd{ChannelID: "other-channel", MessageID: "msg1"})

		if len(h.received) != 0 {
			t.Errorf("handler should not receive reaction events from a non-dev channel, got %v", h.received)
		}
	})

	t.Run("正常系: dev mode設定時は指定チャンネルのリアクションは配送される", func(t *testing.T) {
		r := NewRouter()
		r.SetDevChannelID("dev-channel")
		h := &recordingReactionAddHandler{}
		r.RegisterReactionAddHandler(h)

		r.DispatchReactionAdd(context.Background(), IncomingReactionAdd{ChannelID: "dev-channel", MessageID: "msg1"})

		if len(h.received) != 1 {
			t.Errorf("handler should receive the reaction event from the dev channel, got %v", h.received)
		}
	})
}
