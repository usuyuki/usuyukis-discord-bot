package morph

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome-dict/uni"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
)

// KagomeAnalyzer はusecase/haiku.MorphAnalyzerのkagome実装。
// kagomeへの依存はこのパッケージ内に閉じ込め、usecase層はモーラ数の列のみを受け取る
//
// UniDic辞書は形態素を活用語幹・活用語尾等の細かい単位に分割するため、「一昨日」「最上川」
// のような辞書に登録済みの複合語・固有名詞であっても最短コスト法で分割されてしまい、
// 5-7-5の句切れと単語境界が一致せず判定に失敗することがある。IPA辞書は逆に複合語・固有名詞を
// まとまりとして持つ傾向が強いため、UniDicで川柳・短歌いずれの判定も失敗した場合のみ
// IPA辞書でも解析し直し、そちらで判定が成功すればそちらの結果を採用するフォールバックを行う。
// IPA辞書は実際にフォールバックが発生するまで使われないため、起動時ロードコストを避けるべく
// 初回利用時に遅延ロードする
type KagomeAnalyzer struct {
	primaryName  string
	primary      *tokenizer.Tokenizer
	fallbackName string
	loadFallback func() (*tokenizer.Tokenizer, error)
}

// NewKagomeAnalyzer はUniDic辞書（第一候補）を読み込んだKagomeAnalyzerを生成する。
// IPA辞書（フォールバック）はここではロードせず、初回フォールバック発生時に遅延ロードする
func NewKagomeAnalyzer() (*KagomeAnalyzer, error) {
	primary, err := tokenizer.New(uni.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("morph: failed to create tokenizer (UniDic): %w", err)
	}
	return &KagomeAnalyzer{
		primaryName:  "kagome v2 / UniDic辞書",
		primary:      primary,
		fallbackName: "kagome v2 / IPA辞書",
		loadFallback: sync.OnceValues(func() (*tokenizer.Tokenizer, error) {
			return tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
		}),
	}, nil
}

// AnalyzeWords はtextを形態素解析し、形態素（単語）ごとの表層形とモーラ数を返す。
// UniDic辞書での解析結果が川柳（5-7-5）・短歌（5-7-5-7-7）のいずれの判定にも失敗した場合、
// IPA辞書での解析結果も試し、そちらで判定が成功すればそちらを返す
func (a *KagomeAnalyzer) AnalyzeWords(text string) ([]haiku.Word, error) {
	words := tokenizeWords(a.primary, text)
	if haiku.JudgeAny(words) {
		return words, nil
	}
	// 合計モーラ数が川柳・短歌の定型から大きく外れている場合、辞書を変えて
	// 再解析してもJudgeAnyが真になることはまずないため、IPA辞書での再解析を省略する
	if !haiku.PossibleTotal(words) {
		return words, nil
	}

	fallback, err := a.loadFallback()
	if err != nil {
		return nil, fmt.Errorf("morph: failed to create tokenizer (IPA): %w", err)
	}
	if fbWords := tokenizeWords(fallback, text); haiku.JudgeAny(fbWords) {
		return fbWords, nil
	}
	return words, nil
}

// tokenizeWords はtをtextに対して実行し、形態素ごとの表層形とモーラ数を返す。
// 読みが辞書から取得できない形態素（記号・未知語等）は表層形をそのまま読みとして扱う
func tokenizeWords(t *tokenizer.Tokenizer, text string) []haiku.Word {
	tokens := t.Tokenize(text)
	words := make([]haiku.Word, 0, len(tokens))
	for _, tok := range tokens {
		reading, ok := tok.Pronunciation()
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
	return words
}

// Name は使用している形態素解析器・辞書を表す表示用文字列を返す。
// AnalyzeWordsはメッセージごとにUniDic辞書・IPA辞書のどちらの結果を採用するか動的に切り替わるため、
// 特定の1回の呼び出しでどちらが採用されたかはこの値からは分からない
// （--debugの形態素解析結果に表示される読み方の違いから判別できる）
func (a *KagomeAnalyzer) Name() string {
	return a.primaryName + "（フォールバック: " + a.fallbackName + "）"
}
