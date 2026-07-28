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
}

// Load は環境変数からConfigを読み込む。必須項目が未設定の場合はerrorを返す
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		DiscordBotToken: getenv("DISCORD_BOT_TOKEN"),
		DatabaseURL:     getenv("DATABASE_URL"),
		AdminPort:       orDefault(getenv("ADMIN_PORT"), "8080"),
	}

	if cfg.DiscordBotToken == "" {
		return Config{}, fmt.Errorf("config: DISCORD_BOT_TOKEN must be set")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL must be set")
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
