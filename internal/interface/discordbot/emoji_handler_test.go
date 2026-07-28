package discordbot

import (
	"context"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
	emojiUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/emoji"
)

type fakeEmojiChannelFinder struct {
	ok        bool
	channelID string
}

func (f *fakeEmojiChannelFinder) Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error) {
	return notifychannel.NotifyChannel{ChannelID: f.channelID}, f.ok, nil
}

func TestEmojiHandler_HandleEmojiUpdate(t *testing.T) {
	tests := []struct {
		name       string
		ev         IncomingEmojiUpdate
		channelOK  bool
		wantCalled bool
	}{
		{
			name:       "正常系: 追加絵文字があり通知先が登録済みなら送信する",
			ev:         IncomingEmojiUpdate{GuildID: "g1", AddedEmojiNames: []string{":new:"}},
			channelOK:  true,
			wantCalled: true,
		},
		{
			name:       "異常系: 追加絵文字がなければ送信しない",
			ev:         IncomingEmojiUpdate{GuildID: "g1", AddedEmojiNames: nil},
			channelOK:  true,
			wantCalled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeMessageSender{}
			uc := emojiUC.New(&fakeEmojiChannelFinder{ok: tt.channelOK, channelID: "c1"}, sender)
			h := NewEmojiHandler(uc)

			if err := h.HandleEmojiUpdate(context.Background(), tt.ev); err != nil {
				t.Fatalf("HandleEmojiUpdate() unexpected error = %v", err)
			}
			if sender.called != tt.wantCalled {
				t.Errorf("HandleEmojiUpdate() called = %v, want %v", sender.called, tt.wantCalled)
			}
		})
	}
}
