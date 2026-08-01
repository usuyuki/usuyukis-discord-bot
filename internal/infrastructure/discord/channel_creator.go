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
// カテゴリを指定しないため@everyoneの閲覧権限を含むギルドのデフォルト権限をそのまま継承する
func (c *ChannelCreator) CreateTextChannel(_ context.Context, guildID, name string) (string, error) {
	ch, err := c.session.GuildChannelCreate(guildID, name, discordgo.ChannelTypeGuildText)
	if err != nil {
		return "", fmt.Errorf("discord: failed to create channel %q in guild %s: %w", name, guildID, err)
	}
	return ch.ID, nil
}
