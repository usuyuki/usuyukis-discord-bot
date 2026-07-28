package emoji

import (
	"context"
	"testing"

	domainEmoji "github.com/usuyuki/usuyukis-discord-bot/internal/domain/emoji"
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
	newEmoji := func() domainEmoji.Emoji {
		e, err := domainEmoji.New("new_emoji", "999", false)
		if err != nil {
			t.Fatalf("domainEmoji.New() unexpected error = %v", err)
		}
		return e
	}

	tests := []struct {
		name        string
		addedEmojis []domainEmoji.Emoji
		channelOK   bool
		channelID   string
		wantCalled  bool
		wantContent string
	}{
		{
			name:        "正常系: 通知先が登録済みなら絵文字タグと名前を含む文言で送信する",
			addedEmojis: []domainEmoji.Emoji{newEmoji()},
			channelOK:   true,
			channelID:   "c1",
			wantCalled:  true,
			wantContent: "新しい絵文字が追加されたぱか: <:new_emoji:999> new_emoji",
		},
		{
			name:        "異常系: 追加絵文字がなければ送信しない",
			addedEmojis: []domainEmoji.Emoji{},
			channelOK:   true,
			channelID:   "c1",
			wantCalled:  false,
		},
		{
			name:        "異常系: 通知先が未登録なら送信しない",
			addedEmojis: []domainEmoji.Emoji{newEmoji()},
			channelOK:   false,
			wantCalled:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := &fakeChannelFinder{ok: tt.channelOK, nc: notifychannel.NotifyChannel{ChannelID: tt.channelID}}
			sender := &fakeSender{}
			u := New(finder, sender)

			if err := u.NotifyAdded(context.Background(), "g1", tt.addedEmojis); err != nil {
				t.Fatalf("NotifyAdded() unexpected error = %v", err)
			}
			if sender.called != tt.wantCalled {
				t.Fatalf("NotifyAdded() called = %v, want %v", sender.called, tt.wantCalled)
			}
			if tt.wantCalled && sender.sentChannelID != tt.channelID {
				t.Errorf("NotifyAdded() sentChannelID = %q, want %q", sender.sentChannelID, tt.channelID)
			}
			if tt.wantCalled && sender.sentContent != tt.wantContent {
				t.Errorf("NotifyAdded() sentContent = %q, want %q", sender.sentContent, tt.wantContent)
			}
		})
	}
}
