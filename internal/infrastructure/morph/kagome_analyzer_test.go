package morph

import (
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
)

func TestKagomeAnalyzer_AnalyzeWords(t *testing.T) {
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
		{
			name:      "正常系: UniDicでは一昨と日に分割され575にならないがIPA辞書では一昨日が1語になり575と判定される",
			text:      "落ちたもの一昨日食べて忘れよう",
			wantJudge: true,
		},
		{
			name:      "正常系: UniDicでは最上と川に分割され575にならないがIPA辞書では最上川が1語になり575と判定される",
			text:      "最上川眺め夕焼け染まる空",
			wantJudge: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words, err := analyzer.AnalyzeWords(tt.text)
			if err != nil {
				t.Fatalf("AnalyzeWords() unexpected error = %v", err)
			}
			counts := make([]int, len(words))
			for i, w := range words {
				counts[i] = w.MoraCount
			}
			if got := haiku.Judge(counts); got != tt.wantJudge {
				t.Errorf("Judge(AnalyzeWords(%q)) = %v (counts=%v), want %v", tt.text, got, counts, tt.wantJudge)
			}
		})
	}
}

func TestKagomeAnalyzer_Name(t *testing.T) {
	analyzer, err := NewKagomeAnalyzer()
	if err != nil {
		t.Fatalf("NewKagomeAnalyzer() unexpected error = %v", err)
	}

	if got := analyzer.Name(); got == "" {
		t.Errorf("Name() = %q, want non-empty string", got)
	}
}
