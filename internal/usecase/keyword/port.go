package keyword

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
)

// Repository はキーワードの永続化を担うport。infrastructure層が実装する
type Repository interface {
	Save(ctx context.Context, k keyword.Keyword) error
	Delete(ctx context.Context, guildID, word string) error
	FindByGuild(ctx context.Context, guildID string) ([]keyword.Keyword, error)
}
