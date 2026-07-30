package discordbot

import (
	"context"
	"strings"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/slot"
	slotUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/slot"
)

// slotTriggerPhrase はメンションなしでスロットを起動する固定トリガー文言
const slotTriggerPhrase = "うすゆきスロット"

// SlotHandler は「うすゆきスロット」という発言（メンション不要）を受け、ギルドのカスタム絵文字
// （少なければ標準絵文字）から3つ抽選してスロットを回すハンドラ
type SlotHandler struct {
	uc     *slotUC.UseCase
	sender MessageSender
}

// NewSlotHandler はSlotHandlerを生成する
func NewSlotHandler(uc *slotUC.UseCase, sender MessageSender) *SlotHandler {
	return &SlotHandler{uc: uc, sender: sender}
}

// HandleMessage は本文がちょうど"うすゆきスロット"（前後の空白は無視）に一致すればスロットを回して結果を返信する。
// メンションは不要
func (h *SlotHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if strings.TrimSpace(msg.Content) != slotTriggerPhrase {
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
