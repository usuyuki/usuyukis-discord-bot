package channel

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channel"
)

// UseCase は一般ユーザーからのチャンネル作成依頼に関するアプリケーションロジック。
// Discordの生の「チャンネルを管理」権限を一般ユーザーへ付与すると、新規チャンネル作成だけでなく
// 既存の非公開チャンネルの閲覧・削除・移動まで許してしまう。そのため一般ユーザーには
// その権限を渡さず、Bot自身がこのユースケース経由でチャンネル作成のみを代行し、
// 作成したチャンネルは依頼者以外（サーバー管理者を除く）から見えない状態にする
type UseCase struct {
	creator Creator
}

// New はUseCaseを生成する
func New(creator Creator) *UseCase {
	return &UseCase{creator: creator}
}

// Create はguildIDへrawNameのチャンネルを作成し、依頼者(creatorUserID)以外には
// 見えない状態にした上で作成されたチャンネルIDを返す
func (u *UseCase) Create(ctx context.Context, guildID, rawName, creatorUserID string) (string, error) {
	name, err := channel.NewName(rawName)
	if err != nil {
		return "", err
	}
	return u.creator.CreatePrivateChannel(ctx, guildID, name.String(), creatorUserID)
}
