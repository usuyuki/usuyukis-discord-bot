package channelcreate

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channelcreate"
)

// ChannelCreator はDiscordギルドへの実際のチャンネル作成を担うport。infrastructure層が実装する
type ChannelCreator interface {
	// CreateChannel はguildIDへテキストチャンネルを作成し、作成したチャンネルIDを返す。
	// creation.Privateがtrueの場合、creation.CreatorIDおよびcreation.MemberIDsに含まれる
	// ユーザーのみが閲覧できるチャンネルとして作成する
	CreateChannel(ctx context.Context, guildID string, creation channelcreate.ChannelCreation) (channelID string, err error)
}
