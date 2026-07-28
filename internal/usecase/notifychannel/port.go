package notifychannel

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

// Repository は通知先チャンネル設定の永続化を担うport。infrastructure層が実装する
type Repository interface {
	Set(ctx context.Context, nc notifychannel.NotifyChannel) error
	Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error)
}
