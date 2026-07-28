package keyword

import (
	"errors"
	"strings"
)

var (
	// ErrEmptyKeyword はキーワードが空文字の場合に返す
	ErrEmptyKeyword = errors.New("keyword: keyword must not be empty")
	// ErrEmptyResponse は応答文言が空文字の場合に返す
	ErrEmptyResponse = errors.New("keyword: response must not be empty")
	// ErrEmptyGuildID はギルドIDが空文字の場合に返す
	ErrEmptyGuildID = errors.New("keyword: guildID must not be empty")
)

// Keyword はギルドごとに登録されるキーワード自動応答の値オブジェクト
type Keyword struct {
	ID       int64
	GuildID  string
	Word     string
	Response string
}

// New はKeywordを生成する。空文字のGuildID/Word/Responseは許容しない
func New(id int64, guildID, word, response string) (Keyword, error) {
	if guildID == "" {
		return Keyword{}, ErrEmptyGuildID
	}
	if word == "" {
		return Keyword{}, ErrEmptyKeyword
	}
	if response == "" {
		return Keyword{}, ErrEmptyResponse
	}
	return Keyword{ID: id, GuildID: guildID, Word: word, Response: response}, nil
}

// Matches はメッセージ本文にこのキーワードが部分一致で含まれているかを判定する
func (k Keyword) Matches(messageBody string) bool {
	return strings.Contains(messageBody, k.Word)
}
