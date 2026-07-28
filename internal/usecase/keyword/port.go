package keyword

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
)

// Repository はキーワードの永続化を担うport。infrastructure層が実装する
type Repository interface {
	// AddResponse はギルドの指定キーワードに応答候補を1件追加する。
	// キーワードが未登録なら新規作成し、既存なら応答候補を積み増す
	AddResponse(ctx context.Context, guildID, word, response string) error
	// RemoveResponse はギルドの指定キーワードから特定の応答候補を1件削除する。
	// その応答が最後の1件だった場合はキーワード自体も削除する
	RemoveResponse(ctx context.Context, guildID, word, response string) error
	// RemoveKeyword はギルドから指定キーワードを応答候補ごと全て削除する
	RemoveKeyword(ctx context.Context, guildID, word string) error
	// ReplaceResponses はギルドの指定キーワードの応答候補を渡されたresponsesの内容に丸ごと置き換える
	ReplaceResponses(ctx context.Context, guildID, word string, responses []string) error
	FindByGuild(ctx context.Context, guildID string) ([]keyword.Keyword, error)
}
