package haiku

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

// MorphAnalyzer はテキストを形態素解析し、形態素ごとのモーラ数の列を返すport。
// infrastructure層（kagome実装）が提供する
type MorphAnalyzer interface {
	MoraCountsByWord(text string) ([]int, error)
}

// NotifyChannelFinder はギルド・用途ごとの通知先チャンネルを取得するport
type NotifyChannelFinder interface {
	Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error)
}

// MessageSender はDiscordチャンネルへメッセージを送信するport
type MessageSender interface {
	SendMessage(ctx context.Context, channelID, content string) error
}
