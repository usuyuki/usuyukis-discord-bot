package voicecall

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/voicecall"
)

// UseCase は音声通話の開始・終了通知に関するアプリケーションロジック
type UseCase struct {
	channelFinder NotifyChannelFinder
	sender        EmbedSender
}

// New はUseCaseを生成する
func New(channelFinder NotifyChannelFinder, sender EmbedSender) *UseCase {
	return &UseCase{channelFinder: channelFinder, sender: sender}
}

// NotifyStarted は音声通話の開始を、ギルドに登録された通知先チャンネルへ通知する。
// 通知先が未登録の場合は何もしない（fallback先を持たない仕様）
func (u *UseCase) NotifyStarted(ctx context.Context, guildID string, notice voicecall.StartedNotice) error {
	channelID, ok, err := u.findChannel(ctx, guildID)
	if err != nil || !ok {
		return err
	}
	return u.sender.SendCallStarted(ctx, channelID, notice)
}

// NotifyEnded は音声通話の終了を、ギルドに登録された通知先チャンネルへ通知する。
// 通知先が未登録の場合は何もしない（fallback先を持たない仕様）
func (u *UseCase) NotifyEnded(ctx context.Context, guildID string, notice voicecall.EndedNotice) error {
	channelID, ok, err := u.findChannel(ctx, guildID)
	if err != nil || !ok {
		return err
	}
	return u.sender.SendCallEnded(ctx, channelID, notice)
}

func (u *UseCase) findChannel(ctx context.Context, guildID string) (channelID string, ok bool, err error) {
	nc, ok, err := u.channelFinder.Find(ctx, guildID, notifychannel.PurposeVoiceCall)
	if err != nil || !ok {
		return "", ok, err
	}
	return nc.ChannelID, true, nil
}
