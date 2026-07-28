package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

// NotifyChannelRepository はusecase/notifychannel.Repository（および
// usecase/haiku, usecase/emoji が要求する同形のport）のpgx実装
type NotifyChannelRepository struct {
	pool *pgxpool.Pool
}

// NewNotifyChannelRepository はNotifyChannelRepositoryを生成する
func NewNotifyChannelRepository(pool *pgxpool.Pool) *NotifyChannelRepository {
	return &NotifyChannelRepository{pool: pool}
}

// Set はギルド・用途ごとの通知先チャンネルを設定（upsert）する
func (r *NotifyChannelRepository) Set(ctx context.Context, nc notifychannel.NotifyChannel) error {
	const query = `
		INSERT INTO guild_notify_channels (guild_id, purpose, channel_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (guild_id, purpose) DO UPDATE SET channel_id = EXCLUDED.channel_id
	`
	if _, err := r.pool.Exec(ctx, query, nc.GuildID, string(nc.Purpose), nc.ChannelID); err != nil {
		return fmt.Errorf("postgres: failed to set notify channel: %w", err)
	}
	return nil
}

// Find はギルド・用途ごとの通知先チャンネルを取得する。未設定であればokがfalseになる
func (r *NotifyChannelRepository) Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error) {
	const query = `SELECT guild_id, purpose, channel_id FROM guild_notify_channels WHERE guild_id = $1 AND purpose = $2`
	var gID, p, channelID string
	err := r.pool.QueryRow(ctx, query, guildID, string(purpose)).Scan(&gID, &p, &channelID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return notifychannel.NotifyChannel{}, false, nil
		}
		return notifychannel.NotifyChannel{}, false, fmt.Errorf("postgres: failed to find notify channel: %w", err)
	}
	nc, err := notifychannel.New(gID, notifychannel.Purpose(p), channelID)
	if err != nil {
		return notifychannel.NotifyChannel{}, false, fmt.Errorf("postgres: invalid notify channel row: %w", err)
	}
	return nc, true, nil
}
