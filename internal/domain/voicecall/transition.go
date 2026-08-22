package voicecall

// Transition はボイスチャンネルの在室状態の変化を表す
type Transition string

const (
	// TransitionNone は通知不要な変化（在室者数が変化しても0人↔1人以上をまたがない）を表す
	TransitionNone Transition = "none"
	// TransitionStarted は無人だったチャンネルに人が入り通話が始まったことを表す
	TransitionStarted Transition = "started"
	// TransitionEnded は在室者がいたチャンネルが無人になり通話が終わったことを表す
	TransitionEnded Transition = "ended"
)

// DetectTransition はボイスチャンネル更新前後の在室者数から、通知すべき遷移を判定する。
// 0人から1人以上になった場合はStarted、1人以上から0人になった場合はEndedを返す。
// それ以外（人数の増減はあるが0人をまたがない等）はNoneを返す
func DetectTransition(beforeCount, afterCount int) Transition {
	wasEmpty := beforeCount == 0
	isEmpty := afterCount == 0
	switch {
	case wasEmpty && !isEmpty:
		return TransitionStarted
	case !wasEmpty && isEmpty:
		return TransitionEnded
	default:
		return TransitionNone
	}
}
