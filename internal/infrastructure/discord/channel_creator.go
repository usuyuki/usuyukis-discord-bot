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

// CreateTextChannel はguildIDに公開テキストチャンネルnameを作成し、作成したチャンネルのIDを返す。
// カテゴリを指定しないため@everyoneの閲覧権限を含むギルドのデフォルト権限をそのまま継承するが、
// ギルドの@everyone権限設定次第ではBot自身がそのチャンネルを閲覧・発言できない場合があるため、
// Bot自身（メンバー種別のPermissionOverwrite）に閲覧・発言権限を明示的に付与しておく
func (c *ChannelCreator) CreateTextChannel(_ context.Context, guildID, name string) (string, error) {
	ch, err := c.session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:                 name,
		Type:                 discordgo.ChannelTypeGuildText,
		PermissionOverwrites: c.botPermissionOverwrites(),
	})
	if err != nil {
		return "", fmt.Errorf("discord: failed to create channel %q in guild %s: %w", name, guildID, err)
	}
	return ch.ID, nil
}

// botPermissionOverwrites はBot自身に対する閲覧・発言・履歴閲覧権限のPermissionOverwriteを返す。
// セッションがまだREADYイベントを受信しておらずBotのユーザーIDが分からない場合は空を返し、
// ギルドのデフォルト権限にフォールバックする
func (c *ChannelCreator) botPermissionOverwrites() []*discordgo.PermissionOverwrite {
	if c.session.State == nil || c.session.State.User == nil {
		return nil
	}
	return []*discordgo.PermissionOverwrite{
		{
			ID:   c.session.State.User.ID,
			Type: discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionViewChannel |
				discordgo.PermissionSendMessages |
				discordgo.PermissionReadMessageHistory,
		},
	}
}
