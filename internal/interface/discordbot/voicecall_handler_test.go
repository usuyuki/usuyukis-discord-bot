package discordbot

import (
	"context"
	"testing"
	"time"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/voicecall"
)

type fakeVoiceCallUseCase struct {
	startedCalled    bool
	endedCalled      bool
	gotGuildID       string
	gotStartedNotice voicecall.StartedNotice
	gotEndedNotice   voicecall.EndedNotice
}

func (f *fakeVoiceCallUseCase) NotifyStarted(ctx context.Context, guildID string, notice voicecall.StartedNotice) error {
	f.startedCalled = true
	f.gotGuildID = guildID
	f.gotStartedNotice = notice
	return nil
}

func (f *fakeVoiceCallUseCase) NotifyEnded(ctx context.Context, guildID string, notice voicecall.EndedNotice) error {
	f.endedCalled = true
	f.gotGuildID = guildID
	f.gotEndedNotice = notice
	return nil
}

func TestVoiceCallHandler_HandleVoiceCallStarted(t *testing.T) {
	t.Run("正常系: 開始通知usecaseへイベントの内容がそのまま渡される", func(t *testing.T) {
		uc := &fakeVoiceCallUseCase{}
		h := NewVoiceCallHandler(uc)
		notice := voicecall.StartedNotice{VoiceChannelName: "雑談", StarterName: "usuyuki", StartedAt: time.Date(2026, 6, 9, 1, 52, 51, 0, time.UTC)}

		if err := h.HandleVoiceCallStarted(context.Background(), IncomingVoiceCallStarted{GuildID: "g1", Notice: notice}); err != nil {
			t.Fatalf("HandleVoiceCallStarted() unexpected error = %v", err)
		}
		if !uc.startedCalled {
			t.Fatal("HandleVoiceCallStarted() should call NotifyStarted")
		}
		if uc.gotGuildID != "g1" || uc.gotStartedNotice != notice {
			t.Errorf("HandleVoiceCallStarted() gotGuildID/gotStartedNotice = %q/%+v, want %q/%+v", uc.gotGuildID, uc.gotStartedNotice, "g1", notice)
		}
	})
}

func TestVoiceCallHandler_HandleVoiceCallEnded(t *testing.T) {
	t.Run("正常系: 終了通知usecaseへイベントの内容がそのまま渡される", func(t *testing.T) {
		uc := &fakeVoiceCallUseCase{}
		h := NewVoiceCallHandler(uc)
		notice := voicecall.EndedNotice{VoiceChannelName: "雑談", Duration: 89 * time.Minute, DurationKnown: true}

		if err := h.HandleVoiceCallEnded(context.Background(), IncomingVoiceCallEnded{GuildID: "g1", Notice: notice}); err != nil {
			t.Fatalf("HandleVoiceCallEnded() unexpected error = %v", err)
		}
		if !uc.endedCalled {
			t.Fatal("HandleVoiceCallEnded() should call NotifyEnded")
		}
		if uc.gotGuildID != "g1" || uc.gotEndedNotice != notice {
			t.Errorf("HandleVoiceCallEnded() gotGuildID/gotEndedNotice = %q/%+v, want %q/%+v", uc.gotGuildID, uc.gotEndedNotice, "g1", notice)
		}
	})
}
