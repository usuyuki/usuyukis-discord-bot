package discordbot

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/voicecall"
)

// VoiceCallUseCase はボイスチャンネルの通話開始・終了通知コマンドが呼び出すport
type VoiceCallUseCase interface {
	NotifyStarted(ctx context.Context, guildID string, notice voicecall.StartedNotice) error
	NotifyEnded(ctx context.Context, guildID string, notice voicecall.EndedNotice) error
}

// VoiceCallHandler はボイスチャンネルの在室状態変化イベントを受け、
// 通話の開始・終了を通知usecaseへ橋渡しするハンドラ
type VoiceCallHandler struct {
	uc VoiceCallUseCase
}

// NewVoiceCallHandler はVoiceCallHandlerを生成する
func NewVoiceCallHandler(uc VoiceCallUseCase) *VoiceCallHandler {
	return &VoiceCallHandler{uc: uc}
}

// HandleVoiceCallStarted は通話開始の通知usecaseを呼び出す
func (h *VoiceCallHandler) HandleVoiceCallStarted(ctx context.Context, ev IncomingVoiceCallStarted) error {
	return h.uc.NotifyStarted(ctx, ev.GuildID, ev.Notice)
}

// HandleVoiceCallEnded は通話終了の通知usecaseを呼び出す
func (h *VoiceCallHandler) HandleVoiceCallEnded(ctx context.Context, ev IncomingVoiceCallEnded) error {
	return h.uc.NotifyEnded(ctx, ev.GuildID, ev.Notice)
}
