package channelcreate

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channelcreate"
)

// UseCase はチャンネル作成コマンドに関するアプリケーションロジックをまとめる
type UseCase struct {
	creator ChannelCreator
}

// New はUseCaseを生成する
func New(creator ChannelCreator) *UseCase {
	return &UseCase{creator: creator}
}

// Create は入力を検証した上でguildIDへチャンネルを作成し、作成したチャンネルIDを返す。
// privateがtrueの場合、creatorIDおよびmemberIDsに含まれるユーザーのみが閲覧できるチャンネルとして
// 作成する
func (u *UseCase) Create(ctx context.Context, guildID, name string, private bool, creatorID string, memberIDs []string) (string, error) {
	creation, err := channelcreate.New(name, private, creatorID, memberIDs)
	if err != nil {
		return "", err
	}
	return u.creator.CreateChannel(ctx, guildID, creation)
}
