package morph

import (
	"fmt"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
)

// KagomeAnalyzer はusecase/haiku.MorphAnalyzerのkagome実装。
// kagomeへの依存はこのパッケージ内に閉じ込め、usecase層はモーラ数の列のみを受け取る
type KagomeAnalyzer struct {
	tokenizer *tokenizer.Tokenizer
}

// NewKagomeAnalyzer はIPA辞書を読み込んだKagomeAnalyzerを生成する
func NewKagomeAnalyzer() (*KagomeAnalyzer, error) {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("morph: failed to create tokenizer: %w", err)
	}
	return &KagomeAnalyzer{tokenizer: t}, nil
}

// AnalyzeWords はtextを形態素解析し、形態素（単語）ごとの表層形とモーラ数を返す。
// 読みが辞書から取得できない形態素（記号・未知語等）は表層形をそのまま読みとして扱う
func (a *KagomeAnalyzer) AnalyzeWords(text string) ([]haiku.Word, error) {
	tokens := a.tokenizer.Tokenize(text)
	words := make([]haiku.Word, 0, len(tokens))
	for _, tok := range tokens {
		reading, ok := tok.Reading()
		if !ok || reading == "" {
			reading = tok.Surface
		}
		pos := strings.Join(tok.POS(), "-")
		morae := haiku.SplitMorae(reading)
		if len(morae) == 0 {
			continue
		}
		words = append(words, haiku.Word{
			Surface:   tok.Surface,
			Reading:   reading,
			POS:       pos,
			MoraCount: len(morae),
		})
	}
	return words, nil
}
