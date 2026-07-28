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

// HandleMessage はBotへのメンション（構造化メンションまたは本文中の"@Rakuro"/"@rakuro"表記）を
// 検知したら現在時刻を返信する。
// ただし他ハンドラ（keywordコマンド等）向けのコマンドとの二重応答を避けるため、
// 他コマンドとして解釈できるメッセージには反応しない
func (h *DakokuHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if !msg.MentionsRakuro() {
		return nil
	}
	if parseKeywordCommand(msg.Content) != nil {
		return nil
	}
	reply := dakokuUC.Reply(h.now())
	return h.sender.SendMessage(ctx, msg.ChannelID, reply)
}
