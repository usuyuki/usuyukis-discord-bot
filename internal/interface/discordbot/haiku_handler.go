package discordbot

import (
	"context"

	haikuUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/haiku"
)

// HaikuHandler は投稿を形態素解析して5-7-5（俳句）・5-7-5-7-7（短歌）判定を行い、
// 該当すれば通知するハンドラ
type HaikuHandler struct {
	uc *haikuUC.UseCase
}

// NewHaikuHandler はHaikuHandlerを生成する
func NewHaikuHandler(uc *haikuUC.UseCase) *HaikuHandler {
	return &HaikuHandler{uc: uc}
}

// HandleMessage はBotへのメンションでない通常メッセージを俳句・短歌判定にかける。
// 通知先が未登録の場合は投稿元チャンネルへfallbackする
func (h *HaikuHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if msg.MentionsBotID {
		return nil
	}
	_, err := h.uc.Detect(ctx, msg.GuildID, msg.ChannelID, msg.Content)
	return err
}
