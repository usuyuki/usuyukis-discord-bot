package discord

import (
	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/emoji"
)

// convertEmojis はdiscordgoの絵文字スライスをdomainのemoji.Emojiスライスに変換する。
// 個々の変換に失敗した要素（name/idが空等）は結果から除外する
func convertEmojis(src []*discordgo.Emoji) []emoji.Emoji {
	result := make([]emoji.Emoji, 0, len(src))
	for _, em := range src {
		de, err := emoji.New(em.Name, em.ID, em.Animated)
		if err != nil {
			continue
		}
		result = append(result, de)
	}
	return result
}
