// Package channel はユーザーからの依頼で作成するDiscordチャンネルの名前を表す
// 値オブジェクトを提供する
package channel

import (
	"errors"
	"strings"
)

var (
	// ErrEmptyName はチャンネル名が空文字（前後の空白のみを含む場合を含む）の場合に返す
	ErrEmptyName = errors.New("channel: name must not be empty")
	// ErrNameTooLong はチャンネル名がDiscordの上限文字数を超える場合に返す
	ErrNameTooLong = errors.New("channel: name must be 100 characters or fewer")
)

// maxNameLength はDiscordのチャンネル名に許される上限文字数
const maxNameLength = 100

// Name はDiscord上に作成するチャンネルの名前を表す値オブジェクト。
// 大文字小文字の統一やスペースのハイフン変換はDiscord API側が作成時に自動で行うため、
// ここでは空文字・上限文字数超過の拒否のみを行う
type Name struct {
	value string
}

// NewName はrawを前後の空白を除いた上で検証し、Nameを生成する
func NewName(raw string) (Name, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Name{}, ErrEmptyName
	}
	if len([]rune(trimmed)) > maxNameLength {
		return Name{}, ErrNameTooLong
	}
	return Name{value: trimmed}, nil
}

// String は検証済みのチャンネル名文字列を返す
func (n Name) String() string {
	return n.value
}
