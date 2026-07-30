// Package slot はスロット（絵文字3つ抽選）の結果と役判定を表す純粋な値オブジェクトを提供する
package slot

import "errors"

// ErrInvalidReelCount はリール（抽選された絵文字）が3つちょうどでないときに返るエラー
var ErrInvalidReelCount = errors.New("slot: reels must have exactly 3 items")

// Rank はスロットの役
type Rank int

const (
	// RankMiss はどの絵文字も揃わなかった、はずれ
	RankMiss Rank = iota
	// RankSmall は3つ中2つが一致した、小当たり
	RankSmall
	// RankBig は3つとも一致した、大当たり
	RankBig
)

// Result はスロットの抽選結果
type Result struct {
	reels [3]string
	rank  Rank
}

// NewResult は3つの絵文字表示文字列（Tag等）からResultを生成し、役を自動判定する
func NewResult(reels []string) (Result, error) {
	if len(reels) != 3 {
		return Result{}, ErrInvalidReelCount
	}
	var arr [3]string
	copy(arr[:], reels)
	return Result{reels: arr, rank: judge(arr)}, nil
}

// judge は3つの絵文字表示文字列から役を判定する
func judge(reels [3]string) Rank {
	switch {
	case reels[0] == reels[1] && reels[1] == reels[2]:
		return RankBig
	case reels[0] == reels[1] || reels[1] == reels[2] || reels[0] == reels[2]:
		return RankSmall
	default:
		return RankMiss
	}
}

// Reels は抽選された3つの絵文字表示文字列を返す
func (r Result) Reels() [3]string {
	return r.reels
}

// Rank は判定済みの役を返す
func (r Result) Rank() Rank {
	return r.rank
}
