package discordbot

import (
	"context"

	channelUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/channel"
)

// ChannelHandler はギルドへの新規チャンネル作成イベントを受け、非公開チャンネルであれば
// チャンネル管理ロール（一般ユーザーに付与されたManageChannelsを持つロールでAdministratorは
// 持たないもの）のアクセスを剥奪し、作成者本人とサーバー管理者のみが操作できる状態にするハンドラ。
// 一般ユーザーにチャンネル作成用のロールを渡す運用（生のDiscordロール権限）を前提とし、
// このハンドラはその副作用として生じる「他人の非公開チャンネルまで操作できてしまう」問題を防ぐ
type ChannelHandler struct {
	uc *channelUC.UseCase
}

// NewChannelHandler はChannelHandlerを生成する
func NewChannelHandler(uc *channelUC.UseCase) *ChannelHandler {
	return &ChannelHandler{uc: uc}
}

// HandleChannelCreate はevがプライベートチャンネルであれば、チャンネル管理ロールのアクセスを
// 剥奪し作成者本人のみ操作できる状態にする
func (h *ChannelHandler) HandleChannelCreate(ctx context.Context, ev IncomingChannelCreate) error {
	return h.uc.ProtectIfPrivate(ctx, ev.GuildID, ev.ChannelID, ev.CreatorID, ev.IsPrivate)
}
