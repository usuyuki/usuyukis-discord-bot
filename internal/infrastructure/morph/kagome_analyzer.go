package morph

import (
	"fmt"
	"strings"

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
// IPA辞書でも解析し直し、そちらで判定が成功すればそちらの結果を採用するフォールバックを行う
type KagomeAnalyzer struct {
	primaryName  string
	primary      *tokenizer.Tokenizer
	fallbackName string
	fallback     *tokenizer.Tokenizer
}

// NewKagomeAnalyzer はUniDic辞書（第一候補）とIPA辞書（フォールバック）を読み込んだKagomeAnalyzerを生成する
func NewKagomeAnalyzer() (*KagomeAnalyzer, error) {
	primary, err := tokenizer.New(uni.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("morph: failed to create tokenizer (UniDic): %w", err)
	}
	fallback, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("morph: failed to create tokenizer (IPA): %w", err)
	}
	return &KagomeAnalyzer{
		primaryName:  "kagome v2 / UniDic辞書",
		primary:      primary,
		fallbackName: "kagome v2 / IPA辞書",
		fallback:     fallback,
	}, nil
}

// AnalyzeWords はtextを形態素解析し、形態素（単語）ごとの表層形とモーラ数を返す。
// UniDic辞書での解析結果が川柳（5-7-5）・短歌（5-7-5-7-7）のいずれの判定にも失敗した場合、
// IPA辞書での解析結果も試し、そちらで判定が成功すればそちらを返す
func (a *KagomeAnalyzer) AnalyzeWords(text string) ([]haiku.Word, error) {
	words := tokenizeWords(a.primary, text)
	if judgesAny(words) {
		return words, nil
	}

	if fbWords := tokenizeWords(a.fallback, text); judgesAny(fbWords) {
		return fbWords, nil
	}
	return words, nil
}

// judgesAny はwordsが川柳（5-7-5）・短歌（5-7-5-7-7）のいずれかの判定に成功するかを返す
func judgesAny(words []haiku.Word) bool {
	counts := make([]int, len(words))
	for i, w := range words {
		counts[i] = w.MoraCount
	}
	return haiku.Judge(counts) || haiku.JudgeTanka(counts)
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
