package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	channelUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/channel"
)

// ChannelProposalRepository はusecase/channel.ProposalRepositoryのpgx実装
type ChannelProposalRepository struct {
	pool *pgxpool.Pool
}

// NewChannelProposalRepository はChannelProposalRepositoryを生成する
func NewChannelProposalRepository(pool *pgxpool.Pool) *ChannelProposalRepository {
	return &ChannelProposalRepository{pool: pool}
}

// Save はチャンネル作成提案を1件保存する
func (r *ChannelProposalRepository) Save(ctx context.Context, p channelUC.Proposal) error {
	const query = `
		INSERT INTO channel_create_proposals (guild_id, channel_id, message_id, channel_name, proposer_id)
		VALUES ($1, $2, $3, $4, $5)
	`
	if _, err := r.pool.Exec(ctx, query, p.GuildID, p.ChannelID, p.MessageID, p.ChannelName, p.ProposerID); err != nil {
		return fmt.Errorf("postgres: failed to save channel proposal: %w", err)
	}
	return nil
}

// FindByMessage はchannelID/messageIDに一致する提案を取得する。見つからなければfoundがfalseになる
func (r *ChannelProposalRepository) FindByMessage(ctx context.Context, channelID, messageID string) (channelUC.Proposal, bool, error) {
	const query = `
		SELECT guild_id, channel_id, message_id, channel_name, proposer_id, resolved
		FROM channel_create_proposals
		WHERE channel_id = $1 AND message_id = $2
	`
	var p channelUC.Proposal
	err := r.pool.QueryRow(ctx, query, channelID, messageID).Scan(
		&p.GuildID, &p.ChannelID, &p.MessageID, &p.ChannelName, &p.ProposerID, &p.Resolved,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return channelUC.Proposal{}, false, nil
		}
		return channelUC.Proposal{}, false, fmt.Errorf("postgres: failed to find channel proposal: %w", err)
	}
	return p, true, nil
}

// MarkResolved はchannelID/messageIDに一致する提案を解決済みにする
func (r *ChannelProposalRepository) MarkResolved(ctx context.Context, channelID, messageID string) error {
	const query = `UPDATE channel_create_proposals SET resolved = true WHERE channel_id = $1 AND message_id = $2`
	if _, err := r.pool.Exec(ctx, query, channelID, messageID); err != nil {
		return fmt.Errorf("postgres: failed to mark channel proposal resolved: %w", err)
	}
	return nil
}
