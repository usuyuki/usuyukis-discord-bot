package config

import (
	"fmt"
	"os"
)

// Config はプロセス起動に必要な環境変数をまとめた設定値
type Config struct {
	DiscordBotToken string
	DatabaseURL     string
	AdminPort       string
	// DevMode が真の場合、BotはDevChannelID以外のチャンネルの投稿に一切反応しない
	DevMode      bool
	DevChannelID string
}

// Load は環境変数からConfigを読み込む。必須項目が未設定の場合はerrorを返す
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		DiscordBotToken: getenv("DISCORD_BOT_TOKEN"),
		DatabaseURL:     getenv("DATABASE_URL"),
		AdminPort:       orDefault(getenv("ADMIN_PORT"), "8080"),
		DevMode:         getenv("DEV_MODE") == "true",
		DevChannelID:    getenv("DEV_CHANNEL_ID"),
	}

	if cfg.DiscordBotToken == "" {
		return Config{}, fmt.Errorf("config: DISCORD_BOT_TOKEN must be set")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL must be set")
	}
	if cfg.DevMode && cfg.DevChannelID == "" {
		return Config{}, fmt.Errorf("config: DEV_CHANNEL_ID must be set when DEV_MODE is true")
	}

	return cfg, nil
}

// LoadFromEnv はos.Getenvを使ってConfigを読み込む
func LoadFromEnv() (Config, error) {
	return Load(os.Getenv)
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
