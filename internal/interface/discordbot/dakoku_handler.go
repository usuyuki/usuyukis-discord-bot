package discordbot

import (
	"context"
	"time"

	dakokuUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/dakoku"
)

// MessageSender はDiscordチャンネルへの返信を行うport
type MessageSender interface {
	SendMessage(ctx context.Context, channelID, content string) error
}

// DakokuHandler は@Botメンションで現在時刻を返す打刻Botのハンドラ
type DakokuHandler struct {
	sender MessageSender
	now    func() time.Time
}

// NewDakokuHandler はDakokuHandlerを生成する
func NewDakokuHandler(sender MessageSender) *DakokuHandler {
	return &DakokuHandler{sender: sender, now: time.Now}
}

// HandleMessage はBotへのメンションを検知したら現在時刻を返信する
func (h *DakokuHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if !msg.MentionsBotID {
		return nil
	}
	reply := dakokuUC.Reply(h.now())
	return h.sender.SendMessage(ctx, msg.ChannelID, reply)
}
