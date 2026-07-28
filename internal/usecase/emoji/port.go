package emoji

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

// NotifyChannelFinder はギルド・用途ごとの通知先チャンネルを取得するport
type NotifyChannelFinder interface {
	Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error)
}

// MessageSender はDiscordチャンネルへメッセージを送信するport
type MessageSender interface {
	SendMessage(ctx context.Context, channelID, content string) error
}
