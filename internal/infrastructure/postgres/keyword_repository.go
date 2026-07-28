package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
)

// KeywordRepository はusecase/keyword.Repositoryのpgx実装。
// keywordsテーブル（1ギルド1キーワードにつき1行）とkeyword_responsesテーブル
// （1キーワードにつき複数行の応答候補）に正規化されたスキーマを扱う
type KeywordRepository struct {
	pool *pgxpool.Pool
}

// NewKeywordRepository はKeywordRepositoryを生成する
func NewKeywordRepository(pool *pgxpool.Pool) *KeywordRepository {
	return &KeywordRepository{pool: pool}
}

// AddResponse はギルドの指定キーワードに応答候補を1件追加する。
// キーワードが未登録なら新規作成し、既存なら応答候補を積み増す
func (r *KeywordRepository) AddResponse(ctx context.Context, guildID, word, response string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const upsertKeyword = `
		INSERT INTO keywords (guild_id, keyword)
		VALUES ($1, $2)
		ON CONFLICT (guild_id, keyword) DO UPDATE SET keyword = EXCLUDED.keyword
		RETURNING id
	`
	var keywordID int64
	if err := tx.QueryRow(ctx, upsertKeyword, guildID, word).Scan(&keywordID); err != nil {
		return fmt.Errorf("postgres: failed to upsert keyword: %w", err)
	}

	const insertResponse = `
		INSERT INTO keyword_responses (keyword_id, response)
		VALUES ($1, $2)
		ON CONFLICT (keyword_id, response) DO NOTHING
	`
	if _, err := tx.Exec(ctx, insertResponse, keywordID, response); err != nil {
		return fmt.Errorf("postgres: failed to add keyword response: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: failed to commit transaction: %w", err)
	}
	return nil
}

// RemoveResponse はギルドの指定キーワードから特定の応答候補を1件削除する。
// 応答候補が0件になったキーワードはkeywordsからも削除する
func (r *KeywordRepository) RemoveResponse(ctx context.Context, guildID, word, response string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const deleteResponse = `
		DELETE FROM keyword_responses
		USING keywords
		WHERE keyword_responses.keyword_id = keywords.id
		  AND keywords.guild_id = $1
		  AND keywords.keyword = $2
		  AND keyword_responses.response = $3
	`
	if _, err := tx.Exec(ctx, deleteResponse, guildID, word, response); err != nil {
		return fmt.Errorf("postgres: failed to delete keyword response: %w", err)
	}

	const deleteEmptyKeyword = `
		DELETE FROM keywords
		WHERE guild_id = $1
		  AND keyword = $2
		  AND NOT EXISTS (
		    SELECT 1 FROM keyword_responses WHERE keyword_responses.keyword_id = keywords.id
		  )
	`
	if _, err := tx.Exec(ctx, deleteEmptyKeyword, guildID, word); err != nil {
		return fmt.Errorf("postgres: failed to delete emptied keyword: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: failed to commit transaction: %w", err)
	}
	return nil
}

// RemoveKeyword はギルドから指定キーワードを応答候補ごと全て削除する
func (r *KeywordRepository) RemoveKeyword(ctx context.Context, guildID, word string) error {
	const query = `DELETE FROM keywords WHERE guild_id = $1 AND keyword = $2`
	if _, err := r.pool.Exec(ctx, query, guildID, word); err != nil {
		return fmt.Errorf("postgres: failed to delete keyword: %w", err)
	}
	return nil
}

// ReplaceResponses はギルドのキーワードの応答候補をresponsesの内容に丸ごと置き換える。
// 既存の応答候補を全削除してからresponsesを挿入し直すことで実現する
func (r *KeywordRepository) ReplaceResponses(ctx context.Context, guildID, word string, responses []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const upsertKeyword = `
		INSERT INTO keywords (guild_id, keyword)
		VALUES ($1, $2)
		ON CONFLICT (guild_id, keyword) DO UPDATE SET keyword = EXCLUDED.keyword
		RETURNING id
	`
	var keywordID int64
	if err := tx.QueryRow(ctx, upsertKeyword, guildID, word).Scan(&keywordID); err != nil {
		return fmt.Errorf("postgres: failed to upsert keyword: %w", err)
	}

	const deleteResponses = `DELETE FROM keyword_responses WHERE keyword_id = $1`
	if _, err := tx.Exec(ctx, deleteResponses, keywordID); err != nil {
		return fmt.Errorf("postgres: failed to clear keyword responses: %w", err)
	}

	const insertResponse = `INSERT INTO keyword_responses (keyword_id, response) VALUES ($1, $2)`
	for _, response := range responses {
		if _, err := tx.Exec(ctx, insertResponse, keywordID, response); err != nil {
			return fmt.Errorf("postgres: failed to insert keyword response: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: failed to commit transaction: %w", err)
	}
	return nil
}

// FindByGuild はギルドに登録済みの全キーワードを応答候補付きで返す
func (r *KeywordRepository) FindByGuild(ctx context.Context, guildID string) ([]keyword.Keyword, error) {
	const query = `
		SELECT k.id, k.guild_id, k.keyword, kr.response
		FROM keywords k
		JOIN keyword_responses kr ON kr.keyword_id = k.id
		WHERE k.guild_id = $1
		ORDER BY k.id, kr.id
	`
	rows, err := r.pool.Query(ctx, query, guildID)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to query keywords: %w", err)
	}
	defer rows.Close()

	// 行はkeyword単位でグルーピングされていないため（1キーワード:N応答のJOIN結果）、
	// id順に並んだ行を辿りながら同一idの行が続く間は応答候補を積み増していく
	type row struct {
		id       int64
		guildID  string
		word     string
		response string
	}
	var order []int64
	responsesByID := map[int64][]string{}
	wordByID := map[int64]string{}
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.id, &rr.guildID, &rr.word, &rr.response); err != nil {
			return nil, fmt.Errorf("postgres: failed to scan keyword row: %w", err)
		}
		if _, seen := responsesByID[rr.id]; !seen {
			order = append(order, rr.id)
			wordByID[rr.id] = rr.word
		}
		responsesByID[rr.id] = append(responsesByID[rr.id], rr.response)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows error: %w", err)
	}

	result := make([]keyword.Keyword, 0, len(order))
	for _, id := range order {
		k, err := keyword.New(id, guildID, wordByID[id], responsesByID[id])
		if err != nil {
			return nil, fmt.Errorf("postgres: invalid keyword row: %w", err)
		}
		result = append(result, k)
	}
	return result, nil
}
