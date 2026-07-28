package admin

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

// GuildDirectory はBotが参加中のギルド・チャンネル情報を提供するport。
// infrastructure/discord のGuildCacheが実装する
type GuildDirectory interface {
	ListGuilds() []GuildInfo
	ListTextChannels(guildID string) ([]ChannelInfo, error)
	GuildName(guildID string) string
}
