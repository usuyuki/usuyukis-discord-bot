package haiku

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
)

// MorphAnalyzer はテキストを形態素解析し、形態素ごとの表層形とモーラ数の列を返すport。
// infrastructure層（kagome実装）が提供する
type MorphAnalyzer interface {
	AnalyzeWords(text string) ([]haiku.Word, error)
	// Name は使用している形態素解析器・辞書を表す表示用文字列を返す（例: "kagome v2 / UniDic辞書"）。
	// デバッグ出力でのみ使用し、判定ロジックはこの値に依存しない
	Name() string
}

// MessageSender はDiscordチャンネルへメッセージを送信するport
type MessageSender interface {
	SendMessage(ctx context.Context, channelID, content string) error
}
