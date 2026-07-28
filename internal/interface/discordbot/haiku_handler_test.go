package discordbot

import (
	"context"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
	haikuUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/haiku"
)

type fakeAnalyzer struct{ result []int }

func (f *fakeAnalyzer) MoraCountsByWord(text string) ([]int, error) { return f.result, nil }

type fakeChannelFinder struct{ ok bool }

func (f *fakeChannelFinder) Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error) {
	return notifychannel.NotifyChannel{}, f.ok, nil
}

func TestHaikuHandler_HandleMessage(t *testing.T) {
	tests := []struct {
		name          string
		msg           IncomingMessage
		moraCounts    []int
		wantSentCount int
	}{
		{
			name:          "正常系: メンションでない575投稿は通知を送信する",
			msg:           IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false},
			moraCounts:    []int{5, 7, 5},
			wantSentCount: 1,
		},
		{
			name:          "異常系: Botへのメンションは俳句判定の対象外",
			msg:           IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true},
			moraCounts:    []int{5, 7, 5},
			wantSentCount: 0,
		},
		{
			name:          "異常系: 575にならなければ通知しない",
			msg:           IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false},
			moraCounts:    []int{4, 7, 5},
			wantSentCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeMessageSender{}
			uc := haikuUC.New(&fakeAnalyzer{result: tt.moraCounts}, &fakeChannelFinder{ok: false}, sender)
			h := NewHaikuHandler(uc)

			if err := h.HandleMessage(context.Background(), tt.msg); err != nil {
				t.Fatalf("HandleMessage() unexpected error = %v", err)
			}
			gotCount := 0
			if sender.called {
				gotCount = 1
			}
			if gotCount != tt.wantSentCount {
				t.Errorf("HandleMessage() sent count = %d, want %d", gotCount, tt.wantSentCount)
			}
		})
	}
}
