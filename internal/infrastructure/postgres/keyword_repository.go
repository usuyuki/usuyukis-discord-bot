package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
)

// KeywordRepository はusecase/keyword.Repositoryのpgx実装
type KeywordRepository struct {
	pool *pgxpool.Pool
}

// NewKeywordRepository はKeywordRepositoryを生成する
func NewKeywordRepository(pool *pgxpool.Pool) *KeywordRepository {
	return &KeywordRepository{pool: pool}
}

// Save はキーワードを登録する。同一ギルド・同一キーワードが既にあれば応答文言を上書きする
func (r *KeywordRepository) Save(ctx context.Context, k keyword.Keyword) error {
	const query = `
		INSERT INTO keywords (guild_id, keyword, response)
		VALUES ($1, $2, $3)
		ON CONFLICT (guild_id, keyword) DO UPDATE SET response = EXCLUDED.response
	`
	if _, err := r.pool.Exec(ctx, query, k.GuildID, k.Word, k.Response); err != nil {
		return fmt.Errorf("postgres: failed to save keyword: %w", err)
	}
	return nil
}

// Delete はギルドから指定キーワードを削除する
func (r *KeywordRepository) Delete(ctx context.Context, guildID, word string) error {
	const query = `DELETE FROM keywords WHERE guild_id = $1 AND keyword = $2`
	if _, err := r.pool.Exec(ctx, query, guildID, word); err != nil {
		return fmt.Errorf("postgres: failed to delete keyword: %w", err)
	}
	return nil
}

// FindByGuild はギルドに登録済みの全キーワードを返す
func (r *KeywordRepository) FindByGuild(ctx context.Context, guildID string) ([]keyword.Keyword, error) {
	const query = `SELECT id, guild_id, keyword, response FROM keywords WHERE guild_id = $1 ORDER BY id`
	rows, err := r.pool.Query(ctx, query, guildID)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to query keywords: %w", err)
	}
	defer rows.Close()

	var result []keyword.Keyword
	for rows.Next() {
		var id int64
		var gID, word, response string
		if err := rows.Scan(&id, &gID, &word, &response); err != nil {
			return nil, fmt.Errorf("postgres: failed to scan keyword row: %w", err)
		}
		k, err := keyword.New(id, gID, word, response)
		if err != nil {
			return nil, fmt.Errorf("postgres: invalid keyword row: %w", err)
		}
		result = append(result, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows error: %w", err)
	}
	return result, nil
}
