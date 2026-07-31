package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// ChannelCreator はusecase/channelのCreator portのdiscordgo実装
type ChannelCreator struct {
	session *discordgo.Session
}

// NewChannelCreator はChannelCreatorを生成する
func NewChannelCreator(session *discordgo.Session) *ChannelCreator {
	return &ChannelCreator{session: session}
}

// CreatePrivateChannel はguildIDへテキストチャンネルを作成する。作成と同時に、
// @everyoneロール（IDはguildIDと同一）にはViewChannelを拒否し、creatorUserIDには
// ViewChannel/SendMessages/ReadMessageHistoryを許可する権限オーバーライドを設定するため、
// 作成直後から依頼者以外には見えない状態になる。
// サーバー管理者（Administrator権限保持者）はDiscordの仕様上チャンネル単位の権限
// オーバーライドを常にバイパスするため、ここで管理者向けの許可を明示的に設定する必要はない
func (c *ChannelCreator) CreatePrivateChannel(ctx context.Context, guildID, name, creatorUserID string) (string, error) {
	ch, err := c.session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name: name,
		Type: discordgo.ChannelTypeGuildText,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   guildID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: discordgo.PermissionViewChannel,
			},
			{
				ID:    creatorUserID,
				Type:  discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("discord: failed to create channel: %w", err)
	}
	return ch.ID, nil
}
