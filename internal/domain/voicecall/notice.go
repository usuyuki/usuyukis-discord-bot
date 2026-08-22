package voicecall

import "time"

// StartedNotice は通話開始通知に必要な情報を表す値オブジェクト
type StartedNotice struct {
	VoiceChannelName string
	StarterName      string
	StarterAvatarURL string
	StartedAt        time.Time
}

// EndedNotice は通話終了通知に必要な情報を表す値オブジェクト。
// DurationKnownは、Bot再起動により開始時刻を把握できないまま終了を検知した場合にfalseとなり、
// その場合Durationの値は無視される（プロセスメモリのみで開始時刻を保持する仕様のため、
// 再起動をまたいだ通話は通話時間を計算できない）
type EndedNotice struct {
	VoiceChannelName string
	Duration         time.Duration
	DurationKnown    bool
}
