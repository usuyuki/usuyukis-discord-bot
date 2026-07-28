package emoji

import (
	"context"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

type fakeChannelFinder struct {
	nc notifychannel.NotifyChannel
	ok bool
}

func (f *fakeChannelFinder) Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error) {
	return f.nc, f.ok, nil
}

type fakeSender struct {
	called        bool
	sentChannelID string
	sentContent   string
}

func (f *fakeSender) SendMessage(ctx context.Context, channelID, content string) error {
	f.called = true
	f.sentChannelID = channelID
	f.sentContent = content
	return nil
}

func TestUseCase_NotifyAdded(t *testing.T) {
	tests := []struct {
		name            string
		addedEmojiNames []string
		channelOK       bool
		channelID       string
		wantCalled      bool
	}{
		{
			name:            "正常系: 通知先が登録済みなら追加絵文字名を含めて送信する",
			addedEmojiNames: []string{":new_emoji:"},
			channelOK:       true,
			channelID:       "c1",
			wantCalled:      true,
		},
		{
			name:            "異常系: 追加絵文字がなければ送信しない",
			addedEmojiNames: []string{},
			channelOK:       true,
			channelID:       "c1",
			wantCalled:      false,
		},
		{
			name:            "異常系: 通知先が未登録なら送信しない",
			addedEmojiNames: []string{":new_emoji:"},
			channelOK:       false,
			wantCalled:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := &fakeChannelFinder{ok: tt.channelOK, nc: notifychannel.NotifyChannel{ChannelID: tt.channelID}}
			sender := &fakeSender{}
			u := New(finder, sender)

			if err := u.NotifyAdded(context.Background(), "g1", tt.addedEmojiNames); err != nil {
				t.Fatalf("NotifyAdded() unexpected error = %v", err)
			}
			if sender.called != tt.wantCalled {
				t.Fatalf("NotifyAdded() called = %v, want %v", sender.called, tt.wantCalled)
			}
			if tt.wantCalled && sender.sentChannelID != tt.channelID {
				t.Errorf("NotifyAdded() sentChannelID = %q, want %q", sender.sentChannelID, tt.channelID)
			}
		})
	}
}
