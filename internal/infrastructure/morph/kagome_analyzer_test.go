package morph

import (
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
)

func TestKagomeAnalyzer_MoraCountsByWord(t *testing.T) {
	analyzer, err := NewKagomeAnalyzer()
	if err != nil {
		t.Fatalf("NewKagomeAnalyzer() unexpected error = %v", err)
	}

	tests := []struct {
		name      string
		text      string
		wantJudge bool
	}{
		{
			name:      "正常系: 古池や蛙飛び込む水の音は575と判定される",
			text:      "古池や蛙飛び込む水の音",
			wantJudge: true,
		},
		{
			name:      "異常系: 575にならない普通の文章はfalseと判定される",
			text:      "今日はいい天気ですね",
			wantJudge: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counts, err := analyzer.MoraCountsByWord(tt.text)
			if err != nil {
				t.Fatalf("MoraCountsByWord() unexpected error = %v", err)
			}
			if got := haiku.Judge(counts); got != tt.wantJudge {
				t.Errorf("Judge(MoraCountsByWord(%q)) = %v (counts=%v), want %v", tt.text, got, counts, tt.wantJudge)
			}
		})
	}
}
