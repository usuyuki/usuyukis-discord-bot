package haiku

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
)

var ignorePattern = regexp.MustCompile(`[\p{P}\p{S}\s]+`)

// UseCase は俳句（5-7-5）・短歌（5-7-5-7-7）投稿の検知・通知に関するアプリケーションロジック
type UseCase struct {
	analyzer MorphAnalyzer
	sender   MessageSender
}

// New はUseCaseを生成する
func New(analyzer MorphAnalyzer, sender MessageSender) *UseCase {
	return &UseCase{analyzer: analyzer, sender: sender}
}

// Detect はメッセージ本文を形態素解析して5-7-5（川柳）・5-7-5-7-7（短歌）判定を行い、
// 該当すれば投稿元チャンネルへ通知を送信する。判定結果（川柳・短歌いずれかを検知したか）を返す
func (u *UseCase) Detect(ctx context.Context, guildID, channelID, messageBody string) (bool, error) {
	isDebug := false
	trimmed := strings.TrimSpace(messageBody)
	// iPhoneでは「--」が入力しにくいため、「-debug」でも起動できるようにする
	switch {
	case strings.HasSuffix(trimmed, "--debug"):
		isDebug = true
		messageBody = strings.TrimSpace(strings.TrimSuffix(trimmed, "--debug"))
	case strings.HasSuffix(trimmed, "-debug"):
		isDebug = true
		messageBody = strings.TrimSpace(strings.TrimSuffix(trimmed, "-debug"))
	}

	cleanBody := ignorePattern.ReplaceAllString(messageBody, "")
	words, err := u.analyzer.AnalyzeWords(cleanBody)
	if err != nil {
		return false, err
	}

	label, phrases, ok := judgeAndSplit(words)
	if !ok && !isDebug {
		return false, nil
	}

	var contentBuilder strings.Builder
	if ok {
		contentBuilder.WriteString(label + "を検知しました:\n「" + strings.Join(phrases, "　") + "」")
	} else {
		contentBuilder.WriteString("川柳・短歌として検知できませんでした。")
	}

	if isDebug {
		if !ok {
			formatPattern := func(p []int) string {
				return strings.ReplaceAll(strings.Trim(fmt.Sprint(p), "[]"), " ", ",")
			}
			resultMark := func(matched bool) string {
				if matched {
					return "✅"
				}
				return "❌"
			}
			senryuResult := haiku.EvaluateBestSplit(words, haiku.HaikuPattern)
			tankaResult := haiku.EvaluateBestSplit(words, haiku.TankaPattern)
			contentBuilder.WriteString("\n\n【デバッグ: 字余り・字足らず判定】\n")
			fmt.Fprintf(&contentBuilder, "川柳判定%s: 期待:%s　結果:%s\n", resultMark(intSliceEqual(senryuResult, haiku.HaikuPattern)), formatPattern(haiku.HaikuPattern), formatPattern(senryuResult))
			fmt.Fprintf(&contentBuilder, "短歌判定%s: 期待:%s　結果:%s", resultMark(intSliceEqual(tankaResult, haiku.TankaPattern)), formatPattern(haiku.TankaPattern), formatPattern(tankaResult))
		}

		contentBuilder.WriteString("\n\n【デバッグ: 使用形態素解析器】\n" + u.analyzer.Name())

		contentBuilder.WriteString("\n\n【デバッグ: 形態素解析結果】\n```text\n")
		for _, w := range words {
			fmt.Fprintf(&contentBuilder, "%s\t%s\t%s\t(%d拍)\n", w.Surface, w.Reading, w.POS, w.MoraCount)
		}
		contentBuilder.WriteString("```")
	}

	if err := u.sender.SendMessage(ctx, channelID, contentBuilder.String()); err != nil {
		return false, err
	}
	return ok, nil
}

// judgeAndSplit はwordsを川柳・短歌それぞれのパターンで判定し、該当した方の
// ラベル（"川柳"/"短歌"）と句ごとに分割した文字列を返す。どちらにも該当しなければok=false
func judgeAndSplit(words []haiku.Word) (label string, phrases []string, ok bool) {
	if phrases, ok := haiku.Split(words, haiku.HaikuPattern); ok {
		return "川柳", phrases, true
	}
	if phrases, ok := haiku.Split(words, haiku.TankaPattern); ok {
		return "短歌", phrases, true
	}
	return "", nil, false
}

// intSliceEqual はa, bの要素が順序も含めてすべて一致するかを返す
func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
