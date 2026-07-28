package discordbot

import (
	"context"

	emojiUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/emoji"
)

// EmojiHandler はギルド絵文字更新イベントを受け、追加分を検出して通知するハンドラ
type EmojiHandler struct {
	uc *emojiUC.UseCase
}

// NewEmojiHandler はEmojiHandlerを生成する
func NewEmojiHandler(uc *emojiUC.UseCase) *EmojiHandler {
	return &EmojiHandler{uc: uc}
}

// HandleEmojiUpdate は差分検出済みの追加絵文字リストを通知usecaseへ渡す。
// discordgoのState比較による差分検出自体はこのハンドラを呼び出す側（session配線部分）が行い、
// IncomingEmojiUpdate.AddedEmojis として渡ってくる
func (h *EmojiHandler) HandleEmojiUpdate(ctx context.Context, ev IncomingEmojiUpdate) error {
	return h.uc.NotifyAdded(ctx, ev.GuildID, ev.AddedEmojis)
}
