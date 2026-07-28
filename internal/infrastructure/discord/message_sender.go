package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// MessageSender はusecase層の各MessageSender port（webhookpost, haiku, emoji）のdiscordgo実装
type MessageSender struct {
	session *discordgo.Session
}

// NewMessageSender はMessageSenderを生成する
func NewMessageSender(session *discordgo.Session) *MessageSender {
	return &MessageSender{session: session}
}

// SendMessage は指定チャンネルへcontentを送信する
func (s *MessageSender) SendMessage(ctx context.Context, channelID, content string) error {
	if _, err := s.session.ChannelMessageSend(channelID, content); err != nil {
		return fmt.Errorf("discord: failed to send message: %w", err)
	}
	return nil
}
