package discordbot

import "context"

// ChannelReactionUseCase はリアクション追加をきっかけにチャンネル作成提案の可決判定を行うport
type ChannelReactionUseCase interface {
	RecordReaction(ctx context.Context, channelID, messageID string) error
}

// ReactionHandler はメッセージへのリアクション追加イベントを受け、対象がチャンネル作成提案
// メッセージであれば承認状況を再判定するハンドラ
type ReactionHandler struct {
	uc ChannelReactionUseCase
}

// NewReactionHandler はReactionHandlerを生成する
func NewReactionHandler(uc ChannelReactionUseCase) *ReactionHandler {
	return &ReactionHandler{uc: uc}
}

// HandleReactionAdd はevに対応する提案があれば承認状況を再判定し、
// 必要人数に達していればチャンネルを作成する
func (h *ReactionHandler) HandleReactionAdd(ctx context.Context, ev IncomingReactionAdd) error {
	return h.uc.RecordReaction(ctx, ev.ChannelID, ev.MessageID)
}
