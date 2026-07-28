package notifychannel

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

// UseCase は通知先チャンネル設定に関するアプリケーションロジックをまとめる
type UseCase struct {
	repo Repository
}

// New はUseCaseを生成する
func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

// Set はギルド・用途ごとの通知先チャンネルを設定（upsert）する
func (u *UseCase) Set(ctx context.Context, guildID string, purpose notifychannel.Purpose, channelID string) error {
	nc, err := notifychannel.New(guildID, purpose, channelID)
	if err != nil {
		return err
	}
	return u.repo.Set(ctx, nc)
}

// Get はギルド・用途ごとの通知先チャンネルを取得する。未設定であればokがfalseになる
func (u *UseCase) Get(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error) {
	return u.repo.Find(ctx, guildID, purpose)
}
