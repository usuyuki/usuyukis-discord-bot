package haiku

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
)

// MorphAnalyzer はテキストを形態素解析し、形態素ごとの表層形とモーラ数の列を返すport。
// infrastructure層（kagome実装）が提供する
type MorphAnalyzer interface {
	AnalyzeWords(text string) ([]haiku.Word, error)
}

// MessageSender はDiscordチャンネルへメッセージを送信するport
type MessageSender interface {
	SendMessage(ctx context.Context, channelID, content string) error
}
