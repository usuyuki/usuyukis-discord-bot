package discordbot

import (
	"context"
	"errors"
	"testing"
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
}

func TestRouter_DispatchEmojiUpdate(t *testing.T) {
	t.Run("正常系: 登録済み全ハンドラに絵文字更新イベントが配送される", func(t *testing.T) {
		r := NewRouter()
		h := &recordingEmojiHandler{}
		r.RegisterEmojiUpdateHandler(h)

		ev := IncomingEmojiUpdate{GuildID: "g1", AddedEmojiNames: []string{":new:"}}
		r.DispatchEmojiUpdate(context.Background(), ev)

		if len(h.received) != 1 || h.received[0].GuildID != "g1" {
			t.Errorf("handler did not receive expected event: %v", h.received)
		}
	})
}
