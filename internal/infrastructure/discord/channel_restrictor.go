package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channel"
)

// ChannelRestrictor はusecase/channelのRestrictor portのdiscordgo実装
type ChannelRestrictor struct {
	session *discordgo.Session
}

// NewChannelRestrictor はChannelRestrictorを生成する
func NewChannelRestrictor(session *discordgo.Session) *ChannelRestrictor {
	return &ChannelRestrictor{session: session}
}

// RestrictToCreatorAndAdmins はguildID内でチャンネル管理ロール（ManageChannelsを持ち
// Administratorは持たないロール）を全て洗い出し、channelIDに対してそれらのロールの
// ViewChannel/ManageChannelsを拒否する権限オーバーライドを設定する。
// creatorUserIDが空文字でなければ、そのユーザーへViewChannel/SendMessages/
// ReadMessageHistory/ManageChannelsを明示的に許可するメンバー単位のオーバーライドも設定する
// （メンバー単位のオーバーライドはロール単位のオーバーライドより優先されるため、
// 対象ロールを持つ作成者本人はこの許可により引き続きアクセスできる）。
// サーバー管理者（Administrator権限保持者）はDiscordの仕様上チャンネル単位の
// オーバーライドを常にバイパスするため、明示的な許可設定は不要
func (r *ChannelRestrictor) RestrictToCreatorAndAdmins(ctx context.Context, guildID, channelID, creatorUserID string) error {
	guild, err := r.session.State.Guild(guildID)
	if err != nil {
		guild, err = r.session.Guild(guildID)
		if err != nil {
			return fmt.Errorf("discord: failed to fetch guild %s: %w", guildID, err)
		}
	}

	const deny = discordgo.PermissionViewChannel | discordgo.PermissionManageChannels
	for _, role := range guild.Roles {
		if !channel.IsChannelManagerRole(role.Permissions) {
			continue
		}
		if err := r.session.ChannelPermissionSet(channelID, role.ID, discordgo.PermissionOverwriteTypeRole, 0, deny); err != nil {
			return fmt.Errorf("discord: failed to restrict role %s on channel %s: %w", role.ID, channelID, err)
		}
	}

	if creatorUserID != "" {
		const allow = discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory | discordgo.PermissionManageChannels
		if err := r.session.ChannelPermissionSet(channelID, creatorUserID, discordgo.PermissionOverwriteTypeMember, allow, 0); err != nil {
			return fmt.Errorf("discord: failed to grant creator %s access to channel %s: %w", creatorUserID, channelID, err)
		}
	}

	return nil
}
