// Package emoji はDiscordカスタム絵文字を表す純粋な値オブジェクトを提供する
package emoji

import "errors"

// ErrEmptyName は名前が空文字のときに返るエラー
var ErrEmptyName = errors.New("emoji: name must not be empty")

// ErrEmptyID はIDが空文字のときに返るエラー
var ErrEmptyID = errors.New("emoji: id must not be empty")

// Emoji はDiscordのカスタム絵文字を表す値オブジェクト
type Emoji struct {
	name     string
	id       string
	animated bool
}

// New はEmojiを生成する
func New(name, id string, animated bool) (Emoji, error) {
	if name == "" {
		return Emoji{}, ErrEmptyName
	}
	if id == "" {
		return Emoji{}, ErrEmptyID
	}
	return Emoji{name: name, id: id, animated: animated}, nil
}

// Name は絵文字名を返す
func (e Emoji) Name() string {
	return e.name
}

// Tag はDiscordのメッセージ上で絵文字として表示されるタグ文字列を返す。
// 通常絵文字は "<:name:id>"、アニメーション絵文字は "<a:name:id>" の形式になる
func (e Emoji) Tag() string {
	if e.animated {
		return "<a:" + e.name + ":" + e.id + ">"
	}
	return "<:" + e.name + ":" + e.id + ">"
}
