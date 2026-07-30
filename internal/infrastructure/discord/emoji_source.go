package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/emoji"
)

// EmojiSource はslot usecaseのEmojiSource portをdiscordgo.Session.State経由で実装する
type EmojiSource struct {
	session *discordgo.Session
}

// NewEmojiSource はEmojiSourceを生成する
func NewEmojiSource(session *discordgo.Session) *EmojiSource {
	return &EmojiSource{session: session}
}

// ListEmojiTags は指定ギルドに登録済みのカスタム絵文字をタグ文字列（<:name:id>形式）の
// リストとして返す。ギルドがStateから取得できない場合は空リストを返す
// （slot usecase側で標準絵文字セットへのフォールバックに委ねる）
func (s *EmojiSource) ListEmojiTags(ctx context.Context, guildID string) ([]string, error) {
	g, err := s.session.State.Guild(guildID)
	if err != nil {
		return nil, nil
	}

	tags := make([]string, 0, len(g.Emojis))
	for _, em := range g.Emojis {
		de, err := emoji.New(em.Name, em.ID, em.Animated)
		if err != nil {
			continue
		}
		tags = append(tags, de.Tag())
	}
	return tags, nil
}
