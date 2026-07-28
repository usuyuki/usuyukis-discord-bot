package haiku

import (
	"context"
	"strings"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

// haikuPattern / tankaPattern は句ごとの拍数の列。domain.Splitへ渡す判定パターン
var haikuPattern = []int{5, 7, 5}
var tankaPattern = []int{5, 7, 5, 7, 7}

// UseCase は俳句（5-7-5）・短歌（5-7-5-7-7）投稿の検知・通知に関するアプリケーションロジック
type UseCase struct {
	analyzer      MorphAnalyzer
	channelFinder NotifyChannelFinder
	sender        MessageSender
}

// New はUseCaseを生成する
func New(analyzer MorphAnalyzer, channelFinder NotifyChannelFinder, sender MessageSender) *UseCase {
	return &UseCase{analyzer: analyzer, channelFinder: channelFinder, sender: sender}
}

// Detect はメッセージ本文を形態素解析して5-7-5（俳句）・5-7-5-7-7（短歌）判定を行い、
// 該当すれば通知先チャンネル（未登録の場合はfallbackChannelID）へ、句ごとにスペース区切りで
// 通知を送信する。判定結果（俳句・短歌いずれかを検知したか）を返す
func (u *UseCase) Detect(ctx context.Context, guildID, fallbackChannelID, messageBody string) (bool, error) {
	words, err := u.analyzer.AnalyzeWords(messageBody)
	if err != nil {
		return false, err
	}

	label, phrases, ok := judgeAndSplit(words)
	if !ok {
		return false, nil
	}

	channelID := fallbackChannelID
	nc, found, err := u.channelFinder.Find(ctx, guildID, notifychannel.PurposeHaiku)
	if err != nil {
		return false, err
	}
	if found {
		channelID = nc.ChannelID
	}

	content := label + "を検知しました:\n" + strings.Join(phrases, " ")
	if err := u.sender.SendMessage(ctx, channelID, content); err != nil {
		return false, err
	}
	return true, nil
}

// judgeAndSplit はwordsを俳句・短歌それぞれのパターンで判定し、該当した方の
// ラベル（"俳句"/"短歌"）と句ごとに分割した文字列を返す。どちらにも該当しなければok=false
func judgeAndSplit(words []haiku.Word) (label string, phrases []string, ok bool) {
	if phrases, ok := haiku.Split(words, haikuPattern); ok {
		return "俳句", phrases, true
	}
	if phrases, ok := haiku.Split(words, tankaPattern); ok {
		return "短歌", phrases, true
	}
	return "", nil, false
}
