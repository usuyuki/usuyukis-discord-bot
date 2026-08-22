package voicecall

import (
	"context"
	"testing"
	"time"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/voicecall"
)

type fakeChannelFinder struct {
	nc notifychannel.NotifyChannel
	ok bool
}

func (f *fakeChannelFinder) Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error) {
	return f.nc, f.ok, nil
}

type fakeEmbedSender struct {
	startedCalled     bool
	endedCalled       bool
	sentChannelID     string
	sentStartedNotice voicecall.StartedNotice
	sentEndedNotice   voicecall.EndedNotice
}

func (f *fakeEmbedSender) SendCallStarted(ctx context.Context, channelID string, notice voicecall.StartedNotice) error {
	f.startedCalled = true
	f.sentChannelID = channelID
	f.sentStartedNotice = notice
	return nil
}

func (f *fakeEmbedSender) SendCallEnded(ctx context.Context, channelID string, notice voicecall.EndedNotice) error {
	f.endedCalled = true
	f.sentChannelID = channelID
	f.sentEndedNotice = notice
	return nil
}

func TestUseCase_NotifyStarted(t *testing.T) {
	notice := voicecall.StartedNotice{VoiceChannelName: "雑談", StarterName: "usuyuki", StartedAt: time.Date(2026, 6, 9, 1, 52, 51, 0, time.UTC)}

	tests := []struct {
		name       string
		channelOK  bool
		channelID  string
		wantCalled bool
	}{
		{name: "正常系: 通知先が登録済みならEmbedを送信する", channelOK: true, channelID: "notify-c1", wantCalled: true},
		{name: "異常系: 通知先が未登録なら送信しない", channelOK: false, wantCalled: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := &fakeChannelFinder{ok: tt.channelOK, nc: notifychannel.NotifyChannel{ChannelID: tt.channelID}}
			sender := &fakeEmbedSender{}
			u := New(finder, sender)

			if err := u.NotifyStarted(context.Background(), "g1", notice); err != nil {
				t.Fatalf("NotifyStarted() unexpected error = %v", err)
			}
			if sender.startedCalled != tt.wantCalled {
				t.Fatalf("NotifyStarted() called = %v, want %v", sender.startedCalled, tt.wantCalled)
			}
			if tt.wantCalled {
				if sender.sentChannelID != tt.channelID {
					t.Errorf("NotifyStarted() sentChannelID = %q, want %q", sender.sentChannelID, tt.channelID)
				}
				if sender.sentStartedNotice != notice {
					t.Errorf("NotifyStarted() sentStartedNotice = %+v, want %+v", sender.sentStartedNotice, notice)
				}
			}
		})
	}
}

func TestUseCase_NotifyEnded(t *testing.T) {
	notice := voicecall.EndedNotice{VoiceChannelName: "雑談", Duration: 89 * time.Minute, DurationKnown: true}

	tests := []struct {
		name       string
		channelOK  bool
		channelID  string
		wantCalled bool
	}{
		{name: "正常系: 通知先が登録済みならEmbedを送信する", channelOK: true, channelID: "notify-c1", wantCalled: true},
		{name: "異常系: 通知先が未登録なら送信しない", channelOK: false, wantCalled: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := &fakeChannelFinder{ok: tt.channelOK, nc: notifychannel.NotifyChannel{ChannelID: tt.channelID}}
			sender := &fakeEmbedSender{}
			u := New(finder, sender)

			if err := u.NotifyEnded(context.Background(), "g1", notice); err != nil {
				t.Fatalf("NotifyEnded() unexpected error = %v", err)
			}
			if sender.endedCalled != tt.wantCalled {
				t.Fatalf("NotifyEnded() called = %v, want %v", sender.endedCalled, tt.wantCalled)
			}
			if tt.wantCalled {
				if sender.sentChannelID != tt.channelID {
					t.Errorf("NotifyEnded() sentChannelID = %q, want %q", sender.sentChannelID, tt.channelID)
				}
				if sender.sentEndedNotice != notice {
					t.Errorf("NotifyEnded() sentEndedNotice = %+v, want %+v", sender.sentEndedNotice, notice)
				}
			}
		})
	}
}
