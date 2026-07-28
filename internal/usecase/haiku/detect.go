package haiku

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

// UseCase は俳句（5-7-5）投稿の検知・通知に関するアプリケーションロジック
type UseCase struct {
	analyzer      MorphAnalyzer
	channelFinder NotifyChannelFinder
	sender        MessageSender
}

// New はUseCaseを生成する
func New(analyzer MorphAnalyzer, channelFinder NotifyChannelFinder, sender MessageSender) *UseCase {
	return &UseCase{analyzer: analyzer, channelFinder: channelFinder, sender: sender}
}

// Detect はメッセージ本文を形態素解析して5-7-5判定を行い、該当すれば通知先チャンネル
// （未登録の場合はfallbackChannelID）へ通知を送信する。判定結果（俳句か否か）を返す
func (u *UseCase) Detect(ctx context.Context, guildID, fallbackChannelID, messageBody string) (bool, error) {
	moraCounts, err := u.analyzer.MoraCountsByWord(messageBody)
	if err != nil {
		return false, err
	}

	if !haiku.Judge(moraCounts) {
		return false, nil
	}

	channelID := fallbackChannelID
	nc, ok, err := u.channelFinder.Find(ctx, guildID, notifychannel.PurposeHaiku)
	if err != nil {
		return false, err
	}
	if ok {
		channelID = nc.ChannelID
	}

	if err := u.sender.SendMessage(ctx, channelID, "俳句を検知しました: "+messageBody); err != nil {
		return false, err
	}
	return true, nil
}
