package keyword

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
)

// UseCase はキーワード自動応答に関するアプリケーションロジックをまとめる
type UseCase struct {
	repo Repository
}

// New はUseCaseを生成する
func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

// Register はギルドにキーワードと応答文言を登録する
func (u *UseCase) Register(ctx context.Context, guildID, word, response string) error {
	k, err := keyword.New(0, guildID, word, response)
	if err != nil {
		return err
	}
	return u.repo.Save(ctx, k)
}

// Remove はギルドから指定キーワードを削除する
func (u *UseCase) Remove(ctx context.Context, guildID, word string) error {
	return u.repo.Delete(ctx, guildID, word)
}

// List はギルドに登録済みの全キーワードを返す
func (u *UseCase) List(ctx context.Context, guildID string) ([]keyword.Keyword, error) {
	return u.repo.FindByGuild(ctx, guildID)
}

// Match はメッセージ本文に一致する登録済みキーワードを1件返す。一致がなければokがfalseになる
func (u *UseCase) Match(ctx context.Context, guildID, messageBody string) (keyword.Keyword, bool, error) {
	keywords, err := u.repo.FindByGuild(ctx, guildID)
	if err != nil {
		return keyword.Keyword{}, false, err
	}
	for _, k := range keywords {
		if k.Matches(messageBody) {
			return k, true, nil
		}
	}
	return keyword.Keyword{}, false, nil
}
