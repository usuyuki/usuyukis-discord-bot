package slot

import "context"

// EmojiSource はスロットの抽選対象となるギルドのカスタム絵文字一覧を取得するport。
// 返す文字列はメッセージにそのまま出せる表示形式（Discordの<:name:id>タグ等）であることを期待する
type EmojiSource interface {
	ListEmojiTags(ctx context.Context, guildID string) ([]string, error)
}

// Randomizer はスロットの抽選に使う乱数を提供するport
type Randomizer interface {
	// Intn は[0, n)の範囲の乱数を1つ返す
	Intn(n int) int
}
