package keyword

import (
	"context"
	"time"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
)

// UseCase はキーワード自動応答に関するアプリケーションロジックをまとめる
type UseCase struct {
	repo       Repository
	randomizer Randomizer
}

// New はUseCaseを生成する
func New(repo Repository, randomizer Randomizer) *UseCase {
	return &UseCase{repo: repo, randomizer: randomizer}
}

// Register はギルドのキーワードに応答文言を1件追加する。
// 同一キーワードへの複数回の登録は応答候補の積み増しとして扱われ、Slackのカスタム絵文字応答のように
// 1つのキーワードに複数の応答が紐づく（マッチ時はその中からランダムに1つが選ばれる）
func (u *UseCase) Register(ctx context.Context, guildID, word, response string) error {
	if _, err := keyword.New(0, guildID, word, []string{response}); err != nil {
		return err
	}
	return u.repo.AddResponse(ctx, guildID, word, response)
}

// RemoveResponse はギルドのキーワードから特定の応答候補のみを削除する。
// 最後の1件を削除した場合はキーワード自体が登録一覧から消える
func (u *UseCase) RemoveResponse(ctx context.Context, guildID, word, response string) error {
	return u.repo.RemoveResponse(ctx, guildID, word, response)
}

// RemoveKeyword はギルドから指定キーワードを応答候補ごと全て削除する
func (u *UseCase) RemoveKeyword(ctx context.Context, guildID, word string) error {
	return u.repo.RemoveKeyword(ctx, guildID, word)
}

// SetResponses はギルドのキーワードの応答候補をresponsesの内容に丸ごと置き換える。
// Web管理画面での一括編集（改行区切りテキストエリア）向けのユースケース
func (u *UseCase) SetResponses(ctx context.Context, guildID, word string, responses []string) error {
	if _, err := keyword.New(0, guildID, word, responses); err != nil {
		return err
	}
	return u.repo.ReplaceResponses(ctx, guildID, word, responses)
}

// List はギルドに登録済みの全キーワードを返す
func (u *UseCase) List(ctx context.Context, guildID string) ([]keyword.Keyword, error) {
	return u.repo.FindByGuild(ctx, guildID)
}

// Match はメッセージ本文に一致する登録済みキーワードを探し、一致すれば応答候補からランダムに
// 1つ選んで{$now}等のプレースホルダーを展開した応答文字列を返す。一致がなければokがfalseになる
func (u *UseCase) Match(ctx context.Context, guildID, messageBody string, now time.Time) (response string, ok bool, err error) {
	keywords, err := u.repo.FindByGuild(ctx, guildID)
	if err != nil {
		return "", false, err
	}
	for _, k := range keywords {
		if k.Matches(messageBody) {
			return k.ResponseAt(u.randomizer.Intn(len(k.Responses)), now), true, nil
		}
	}
	return "", false, nil
}
