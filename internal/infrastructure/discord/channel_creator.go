package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channelcreate"
)

// ChannelCreator はusecase/channelcreateのChannelCreator portのdiscordgo実装。
// チャンネル作成自体はBotのManage Channels権限で行うため、コマンド実行者は個別の
// チャンネル管理権限を持っている必要がない
type ChannelCreator struct {
	session *discordgo.Session
}

// NewChannelCreator はChannelCreatorを生成する
func NewChannelCreator(session *discordgo.Session) *ChannelCreator {
	return &ChannelCreator{session: session}
}

// CreateChannel はguildIDへテキストチャンネルを作成する。creation.Privateがtrueの場合、
// @everyoneロールに対しチャンネル閲覧権限を拒否した上で、creation.CreatorIDと
// creation.MemberIDsに対してのみ閲覧権限を許可するPermissionOverwriteを設定する
func (c *ChannelCreator) CreateChannel(ctx context.Context, guildID string, creation channelcreate.ChannelCreation) (string, error) {
	data := discordgo.GuildChannelCreateData{
		Name: creation.Name,
		Type: discordgo.ChannelTypeGuildText,
	}
	if creation.Private {
		data.PermissionOverwrites = privateChannelOverwrites(guildID, creation)
	}

	ch, err := c.session.GuildChannelCreateComplex(guildID, data, discordgo.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("discord: failed to create channel: %w", err)
	}
	return ch.ID, nil
}

// privateChannelOverwrites は@everyoneへの閲覧拒否と、作成者・許可メンバーへの閲覧許可からなる
// PermissionOverwrite一覧を組み立てる。@everyoneロールのIDはguildIDと一致するというDiscordの
// 仕様を利用している
func privateChannelOverwrites(guildID string, creation channelcreate.ChannelCreation) []*discordgo.PermissionOverwrite {
	overwrites := []*discordgo.PermissionOverwrite{
		{
			ID:   guildID,
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionViewChannel,
		},
	}

	allowedMembers := append([]string{creation.CreatorID}, creation.MemberIDs...)
	for _, memberID := range allowedMembers {
		overwrites = append(overwrites, &discordgo.PermissionOverwrite{
			ID:    memberID,
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionViewChannel,
		})
	}
	return overwrites
}
