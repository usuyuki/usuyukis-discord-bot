package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// NewSession はBotトークンからdiscordgoセッションを生成し、必要なIntentsを設定する。
// メッセージ本文の取得（メンション・キーワード検知）とギルド絵文字更新の購読に
// MessageContent, Guilds, GuildMessages, GuildEmojis の各Intentを要求する
func NewSession(token string) (*discordgo.Session, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("discord: failed to create session: %w", err)
	}
	s.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentMessageContent |
		discordgo.IntentsGuildEmojis
	return s, nil
}
