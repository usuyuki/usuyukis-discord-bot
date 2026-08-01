package discord

import (
	"context"
	"fmt"
	"log"

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

// CreateTextChannel はguildIDに公開テキストチャンネルnameを作成し、作成したチャンネルへの
// Discordメンション形式（<#channelID>、クライアント側でクリック可能なリンクとして表示される）の
// 参照文字列を返す。カテゴリを指定しないため@everyoneの閲覧権限を含むギルドのデフォルト権限を
// そのまま継承するが、ギルドの@everyone権限設定次第ではBot自身がそのチャンネルを閲覧・発言
// できない場合があるため、Bot自身（メンバー種別のPermissionOverwrite）に閲覧・発言権限を
// 明示的に付与しておく
func (c *ChannelCreator) CreateTextChannel(_ context.Context, guildID, name string) (string, error) {
	ch, err := c.session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:                 name,
		Type:                 discordgo.ChannelTypeGuildText,
		PermissionOverwrites: c.botPermissionOverwrites(),
	})
	if err != nil {
		return "", fmt.Errorf("discord: failed to create channel %q in guild %s: %w", name, guildID, err)
	}
	if ch == nil || ch.ID == "" {
		return "", fmt.Errorf("discord: channel creation returned no channel for %q in guild %s", name, guildID)
	}
	return fmt.Sprintf("<#%s>", ch.ID), nil
}

// botPermissionOverwrites はBot自身に対する閲覧・発言・履歴閲覧権限のPermissionOverwriteを返す。
// セッションがまだREADYイベントを受信しておらずBotのユーザーIDが分からない場合は空を返し、
// ギルドのデフォルト権限にフォールバックする
func (c *ChannelCreator) botPermissionOverwrites() []*discordgo.PermissionOverwrite {
	if c.session.State == nil || c.session.State.User == nil {
		log.Printf("discord: bot user ID unavailable, falling back to guild default permissions for new channel")
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
