package discordbot

import (
	"context"
	"strings"
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

// mentionsRakuroText は本文に"@Rakuro"または"@rakuro"という文字列が含まれるかを
// 大文字小文字を無視して判定する。discordgoの構造化メンション（MentionsBotID）ではなく、
// 本文中のテキストとしての@Rakuro表記のみを打刻Botの起動条件とするための判定
func mentionsRakuroText(content string) bool {
	return strings.Contains(strings.ToLower(content), "@rakuro")
}

// HandleMessage は本文に"@Rakuro"/"@rakuro"という文字列を検知したら現在時刻を返信する。
// ただし他ハンドラ（keywordコマンド等）向けのコマンドとの二重応答を避けるため、
// 他コマンドとして解釈できるメッセージには反応しない
func (h *DakokuHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if !mentionsRakuroText(msg.Content) {
		return nil
	}
	if parseKeywordCommand(msg.Content) != nil {
		return nil
	}
	reply := dakokuUC.Reply(h.now())
	return h.sender.SendMessage(ctx, msg.ChannelID, reply)
}
