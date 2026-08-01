package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ChannelSettingRepository はusecase/channel.SettingRepositoryのpgx実装
type ChannelSettingRepository struct {
	pool *pgxpool.Pool
}

// NewChannelSettingRepository はChannelSettingRepositoryを生成する
func NewChannelSettingRepository(pool *pgxpool.Pool) *ChannelSettingRepository {
	return &ChannelSettingRepository{pool: pool}
}

// Get はguildIDに設定された必要承認人数を取得する。未設定であればfoundがfalseになる
func (r *ChannelSettingRepository) Get(ctx context.Context, guildID string) (int, bool, error) {
	const query = `SELECT required_approvals FROM guild_channel_create_settings WHERE guild_id = $1`
	var required int
	err := r.pool.QueryRow(ctx, query, guildID).Scan(&required)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("postgres: failed to get channel create setting: %w", err)
	}
	return required, true, nil
}

// Set はguildIDの必要承認人数を設定（upsert）する
func (r *ChannelSettingRepository) Set(ctx context.Context, guildID string, requiredApprovals int) error {
	const query = `
		INSERT INTO guild_channel_create_settings (guild_id, required_approvals)
		VALUES ($1, $2)
		ON CONFLICT (guild_id) DO UPDATE SET required_approvals = EXCLUDED.required_approvals
	`
	if _, err := r.pool.Exec(ctx, query, guildID, requiredApprovals); err != nil {
		return fmt.Errorf("postgres: failed to set channel create setting: %w", err)
	}
	return nil
}
