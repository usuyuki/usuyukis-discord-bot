package haiku

import (
	"context"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

type fakeAnalyzer struct {
	result []int
	err    error
}

func (f *fakeAnalyzer) MoraCountsByWord(text string) ([]int, error) {
	return f.result, f.err
}

type fakeChannelFinder struct {
	nc notifychannel.NotifyChannel
	ok bool
}

func (f *fakeChannelFinder) Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error) {
	return f.nc, f.ok, nil
}

type fakeSender struct {
	sentChannelID string
	sentContent   string
	called        bool
}

func (f *fakeSender) SendMessage(ctx context.Context, channelID, content string) error {
	f.sentChannelID = channelID
	f.sentContent = content
	f.called = true
	return nil
}

func TestUseCase_Detect(t *testing.T) {
	tests := []struct {
		name              string
		moraCounts        []int
		notifyChannelOK   bool
		notifyChannelID   string
		fallbackChannelID string
		wantDetected      bool
		wantSentChannelID string
	}{
		{
			name:              "正常系: 575判定でtrueかつ通知先チャンネル登録済みならそこに送信する",
			moraCounts:        []int{5, 7, 5},
			notifyChannelOK:   true,
			notifyChannelID:   "notify-channel",
			fallbackChannelID: "fallback-channel",
			wantDetected:      true,
			wantSentChannelID: "notify-channel",
		},
		{
			name:              "正常系: 通知先未登録ならfallbackチャンネルに送信する",
			moraCounts:        []int{5, 7, 5},
			notifyChannelOK:   false,
			fallbackChannelID: "fallback-channel",
			wantDetected:      true,
			wantSentChannelID: "fallback-channel",
		},
		{
			name:              "異常系: 575判定がfalseなら送信しない",
			moraCounts:        []int{4, 7, 5},
			notifyChannelOK:   false,
			fallbackChannelID: "fallback-channel",
			wantDetected:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &fakeAnalyzer{result: tt.moraCounts}
			finder := &fakeChannelFinder{ok: tt.notifyChannelOK, nc: notifychannel.NotifyChannel{ChannelID: tt.notifyChannelID}}
			sender := &fakeSender{}
			u := New(analyzer, finder, sender)

			detected, err := u.Detect(context.Background(), "g1", tt.fallbackChannelID, "テストメッセージ")
			if err != nil {
				t.Fatalf("Detect() unexpected error = %v", err)
			}
			if detected != tt.wantDetected {
				t.Fatalf("Detect() = %v, want %v", detected, tt.wantDetected)
			}
			if tt.wantDetected {
				if !sender.called {
					t.Fatal("Detect() expected SendMessage to be called, but it wasn't")
				}
				if sender.sentChannelID != tt.wantSentChannelID {
					t.Errorf("Detect() sentChannelID = %q, want %q", sender.sentChannelID, tt.wantSentChannelID)
				}
			} else if sender.called {
				t.Error("Detect() expected SendMessage not to be called, but it was")
			}
		})
	}
}
