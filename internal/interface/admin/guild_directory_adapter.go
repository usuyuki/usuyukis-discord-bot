package admin

import "github.com/usuyuki/usuyukis-discord-bot/internal/infrastructure/discord"

// discordGuildDirectory はinfrastructure/discord.GuildCacheをGuildDirectory portへ適合させるアダプタ。
// クリーンアーキテクチャの依存方向（interface→infrastructure）に従い、
// infrastructure側の型をinterface側で変換する
type discordGuildDirectory struct {
	cache *discord.GuildCache
}

// NewDiscordGuildDirectory はGuildCacheをラップしたGuildDirectoryを生成する
func NewDiscordGuildDirectory(cache *discord.GuildCache) GuildDirectory {
	return &discordGuildDirectory{cache: cache}
}

func (a *discordGuildDirectory) ListGuilds() []GuildInfo {
	guilds := a.cache.ListGuilds()
	result := make([]GuildInfo, 0, len(guilds))
	for _, g := range guilds {
		result = append(result, GuildInfo{ID: g.ID, Name: g.Name})
	}
	return result
}

func (a *discordGuildDirectory) ListTextChannels(guildID string) ([]ChannelInfo, error) {
	channels, err := a.cache.ListTextChannels(guildID)
	if err != nil {
		return nil, err
	}
	result := make([]ChannelInfo, 0, len(channels))
	for _, c := range channels {
		result = append(result, ChannelInfo{ID: c.ID, Name: c.Name})
	}
	return result, nil
}

func (a *discordGuildDirectory) GuildName(guildID string) string {
	return a.cache.GuildName(guildID)
}
