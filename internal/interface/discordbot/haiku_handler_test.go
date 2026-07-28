package discordbot

import (
	"context"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
	haikuUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/haiku"
)

type fakeAnalyzer struct{ result []haiku.Word }

func (f *fakeAnalyzer) AnalyzeWords(text string) ([]haiku.Word, error) { return f.result, nil }

func TestHaikuHandler_HandleMessage(t *testing.T) {
	haikuWords := []haiku.Word{
		{Surface: "ふる", MoraCount: 5},
		{Surface: "いけや", MoraCount: 7},
		{Surface: "かわず", MoraCount: 5},
	}
	notHaikuWords := []haiku.Word{
		{Surface: "ふるい", MoraCount: 4},
		{Surface: "けや", MoraCount: 7},
		{Surface: "かわず", MoraCount: 5},
	}

	tests := []struct {
		name          string
		msg           IncomingMessage
		words         []haiku.Word
		wantSentCount int
	}{
		{
			name:          "正常系: メンションでない575投稿は通知を送信する",
			msg:           IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false},
			words:         haikuWords,
			wantSentCount: 1,
		},
		{
			name:          "異常系: Botへのメンションは川柳判定の対象外",
			msg:           IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: true},
			words:         haikuWords,
			wantSentCount: 0,
		},
		{
			name:          "異常系: 構造化メンションでなくテキストの@Rakuro表記でも川柳判定の対象外（打刻Botとの二重応答防止）",
			msg:           IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false, Content: "@Rakuro ふるいけやかわず"},
			words:         haikuWords,
			wantSentCount: 0,
		},
		{
			name:          "異常系: 575にならなければ通知しない",
			msg:           IncomingMessage{GuildID: "g1", ChannelID: "c1", MentionsBotID: false},
			words:         notHaikuWords,
			wantSentCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeMessageSender{}
			uc := haikuUC.New(&fakeAnalyzer{result: tt.words}, sender)
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
