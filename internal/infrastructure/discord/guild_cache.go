package discord

import "github.com/bwmarrin/discordgo"

// GuildInfo は管理画面表示用のギルド情報
type GuildInfo struct {
	ID   string
	Name string
}

// ChannelInfo は管理画面表示用のチャンネル情報
type ChannelInfo struct {
	ID   string
	Name string
}

// GuildCache はdiscordgoのState経由でギルド・チャンネルの一覧や名前解決を行う
type GuildCache struct {
	session *discordgo.Session
}

// NewGuildCache はGuildCacheを生成する
func NewGuildCache(session *discordgo.Session) *GuildCache {
	return &GuildCache{session: session}
}

// ListGuilds はBotが参加中の全ギルドを返す
func (c *GuildCache) ListGuilds() []GuildInfo {
	guilds := c.session.State.Guilds
	result := make([]GuildInfo, 0, len(guilds))
	for _, g := range guilds {
		result = append(result, GuildInfo{ID: g.ID, Name: g.Name})
	}
	return result
}

// ListTextChannels は指定ギルドのテキストチャンネル一覧を返す
func (c *GuildCache) ListTextChannels(guildID string) ([]ChannelInfo, error) {
	g, err := c.session.State.Guild(guildID)
	if err != nil {
		return nil, err
	}
	result := make([]ChannelInfo, 0, len(g.Channels))
	for _, ch := range g.Channels {
		if ch.Type != discordgo.ChannelTypeGuildText {
			continue
		}
		result = append(result, ChannelInfo{ID: ch.ID, Name: ch.Name})
	}
	return result, nil
}

// GuildName は指定ギルドIDの表示名を返す。取得できなければIDをそのまま返す
func (c *GuildCache) GuildName(guildID string) string {
	g, err := c.session.State.Guild(guildID)
	if err != nil {
		return guildID
	}
	return g.Name
}
