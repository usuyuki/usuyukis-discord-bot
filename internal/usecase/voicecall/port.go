package voicecall

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/voicecall"
)

// NotifyChannelFinder はギルド・用途ごとの通知先チャンネルを取得するport
type NotifyChannelFinder interface {
	Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error)
}

// EmbedSender は音声通話の開始・終了をEmbed形式でDiscordチャンネルへ送信するport。
// Embedの組み立て（色・レイアウトなど）はdiscordgoに依存するためinfrastructure層が担う
type EmbedSender interface {
	SendCallStarted(ctx context.Context, channelID string, notice voicecall.StartedNotice) error
	SendCallEnded(ctx context.Context, channelID string, notice voicecall.EndedNotice) error
}
