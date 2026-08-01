// Package channel はDiscordのチャンネル名に関する純粋な検証ロジックを提供する
package channel

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

// maxNameLength はDiscordのチャンネル名の最大文字数
const maxNameLength = 100

// validNamePattern はDiscordが実際に許容する文字集合を全て検証するものではないが、
// 一般ユーザーが誤って紛らわしい名前（絵文字や制御文字など）を作らないよう、
// 半角英数字・ハイフン・アンダースコア・日本語（ひらがな・カタカナ・漢字）のみに絞る運用上の制約
var validNamePattern = regexp.MustCompile(`^[a-z0-9_\-\p{Hiragana}\p{Katakana}\p{Han}ー]+$`)

// Name はバリデーション済みのチャンネル名を表す値オブジェクト
type Name struct {
	value string
}

// NewName はrawをチャンネル名として検証し、Nameを返す。
// 空文字・最大長超過・許容文字以外を含む場合はエラーを返す
func NewName(raw string) (Name, error) {
	if raw == "" {
		return Name{}, fmt.Errorf("channel: name must not be empty")
	}
	if length := utf8.RuneCountInString(raw); length > maxNameLength {
		return Name{}, fmt.Errorf("channel: name must be %d characters or fewer, got %d", maxNameLength, length)
	}
	if !validNamePattern.MatchString(raw) {
		return Name{}, fmt.Errorf("channel: name %q must contain only lowercase letters, numbers, hyphens, underscores, or Japanese characters", raw)
	}
	return Name{value: raw}, nil
}

// String はチャンネル名の文字列表現を返す
func (n Name) String() string {
	return n.value
}
