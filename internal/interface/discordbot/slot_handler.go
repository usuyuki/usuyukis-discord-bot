package discordbot

import (
	"context"
	"strings"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/slot"
	slotUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/slot"
)

// SlotHandler は「@Bot slot」コマンドを受け、ギルドのカスタム絵文字（少なければ標準絵文字）から
// 3つ抽選してスロットを回すハンドラ
type SlotHandler struct {
	uc     *slotUC.UseCase
	sender MessageSender
}

// NewSlotHandler はSlotHandlerを生成する
func NewSlotHandler(uc *slotUC.UseCase, sender MessageSender) *SlotHandler {
	return &SlotHandler{uc: uc, sender: sender}
}

// HandleMessage はBotへの構造化メンションで「slot」に一致すればスロットを回して結果を返信する
func (h *SlotHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if !msg.MentionsBotID {
		return nil
	}
	filtered := stripMentionTokens(strings.Fields(msg.Content), msg.BotID)
	if len(filtered) == 0 || filtered[0] != "slot" {
		return nil
	}

	result, err := h.uc.Spin(ctx, msg.GuildID)
	if err != nil {
		return err
	}
	return h.sender.SendMessage(ctx, msg.ChannelID, formatSlotResult(result))
}

// formatSlotResult はスロットの抽選結果を通知文言に整形する
func formatSlotResult(result slot.Result) string {
	reels := result.Reels()
	line := strings.Join(reels[:], " | ")
	switch result.Rank() {
	case slot.RankBig:
		return line + "\n🎉 大当たり！"
	case slot.RankSmall:
		return line + "\n✨ 小当たり"
	default:
		return line + "\nはずれ"
	}
}
