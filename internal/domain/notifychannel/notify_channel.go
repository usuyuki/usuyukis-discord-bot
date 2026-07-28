package notifychannel

import "errors"

// Purpose は通知先チャンネルの用途を表す
type Purpose string

const (
	// PurposeHaiku は俳句検知通知の用途
	PurposeHaiku Purpose = "haiku"
	// PurposeEmoji は絵文字追加通知の用途
	PurposeEmoji Purpose = "emoji"
)

// IsValid はPurposeが定義済みの値かどうかを判定する
func (p Purpose) IsValid() bool {
	switch p {
	case PurposeHaiku, PurposeEmoji:
		return true
	default:
		return false
	}
}

var (
	// ErrEmptyGuildID はギルドIDが空文字の場合に返す
	ErrEmptyGuildID = errors.New("notifychannel: guildID must not be empty")
	// ErrEmptyChannelID はチャンネルIDが空文字の場合に返す
	ErrEmptyChannelID = errors.New("notifychannel: channelID must not be empty")
	// ErrInvalidPurpose は未定義のPurposeが渡された場合に返す
	ErrInvalidPurpose = errors.New("notifychannel: purpose is invalid")
)

// NotifyChannel はギルド・用途ごとの通知先チャンネル設定を表す値オブジェクト
type NotifyChannel struct {
	GuildID   string
	Purpose   Purpose
	ChannelID string
}

// New はNotifyChannelを生成する
func New(guildID string, purpose Purpose, channelID string) (NotifyChannel, error) {
	if guildID == "" {
		return NotifyChannel{}, ErrEmptyGuildID
	}
	if !purpose.IsValid() {
		return NotifyChannel{}, ErrInvalidPurpose
	}
	if channelID == "" {
		return NotifyChannel{}, ErrEmptyChannelID
	}
	return NotifyChannel{GuildID: guildID, Purpose: purpose, ChannelID: channelID}, nil
}
