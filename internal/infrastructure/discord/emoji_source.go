package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
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
// リストとして返す
func (s *EmojiSource) ListEmojiTags(ctx context.Context, guildID string) ([]string, error) {
	g, err := s.session.State.Guild(guildID)
	if err != nil {
		return nil, err
	}

	emojis := convertEmojis(g.Emojis)
	tags := make([]string, 0, len(emojis))
	for _, de := range emojis {
		tags = append(tags, de.Tag())
	}
	return tags, nil
}
