package keyword

import (
	"errors"
	"math/rand/v2"
	"strings"
	"time"
)

var (
	// ErrEmptyKeyword はキーワードが空文字の場合に返す
	ErrEmptyKeyword = errors.New("keyword: keyword must not be empty")
	// ErrEmptyResponse は応答文言が1件もない、または空文字が含まれる場合に返す
	ErrEmptyResponse = errors.New("keyword: response must not be empty")
	// ErrEmptyGuildID はギルドIDが空文字の場合に返す
	ErrEmptyGuildID = errors.New("keyword: guildID must not be empty")
)

// nowPlaceholder は応答文言中に含まれていると現在日時に展開されるプレースホルダー。
// 打刻Bot（`@Rakuro`固定の時刻返信）を廃止し、この記法を使ったキーワード自動応答に統一する
const nowPlaceholder = "{$now}"

// nowLayout はnowPlaceholder展開時の日時表示フォーマット
const nowLayout = "2006-01-02 15:04:05"

// jst は日本標準時（UTC+9）。実行環境にtzdataが無くても解決できるようFixedZoneで定義する
var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// Keyword はギルドごとに登録されるキーワード自動応答の値オブジェクト。
// 1つのキーワードに対して複数の応答候補（Responses）を持ち、マッチ時にはその中からランダムに1件選ばれる
type Keyword struct {
	ID        int64
	GuildID   string
	Word      string
	Responses []string
}

// New はKeywordを生成する。空文字のGuildID/Wordや、空のResponses・空文字を含むResponsesは許容しない
func New(id int64, guildID, word string, responses []string) (Keyword, error) {
	if guildID == "" {
		return Keyword{}, ErrEmptyGuildID
	}
	if word == "" {
		return Keyword{}, ErrEmptyKeyword
	}
	if len(responses) == 0 {
		return Keyword{}, ErrEmptyResponse
	}
	for _, r := range responses {
		if r == "" {
			return Keyword{}, ErrEmptyResponse
		}
	}
	return Keyword{ID: id, GuildID: guildID, Word: word, Responses: responses}, nil
}

// Matches はメッセージ本文にこのキーワードが部分一致で含まれているかを判定する
func (k Keyword) Matches(messageBody string) bool {
	return strings.Contains(messageBody, k.Word)
}

// RandomResponse は登録済みの応答候補からランダムに1件選び、含まれるnowPlaceholder（{$now}）を
// nowのJST日時文字列に展開して返す
func (k Keyword) RandomResponse(now time.Time) string {
	response := k.Responses[rand.IntN(len(k.Responses))]
	return strings.ReplaceAll(response, nowPlaceholder, now.In(jst).Format(nowLayout))
}
