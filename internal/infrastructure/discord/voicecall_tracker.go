package discord

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/voicecall"
)

// voiceCallTracker はギルド×ボイスチャンネルごとの在室者数と通話開始時刻を
// プロセスローカルに保持し、VoiceStateUpdateイベントから通話の開始/終了を検知する。
// discordgoのStateはハンドラ呼び出し時点で既に更新後の状態になっており「更新前の人数」を
// 復元できないため、絵文字通知の差分検出と同様に自前でカウンタを持つ
type voiceCallTracker struct {
	mu        sync.Mutex
	occupants map[string]map[string]int       // guildID -> channelID -> 在室者数
	startedAt map[string]map[string]time.Time // guildID -> channelID -> 通話開始時刻
}

func newVoiceCallTracker() *voiceCallTracker {
	return &voiceCallTracker{
		occupants: map[string]map[string]int{},
		startedAt: map[string]map[string]time.Time{},
	}
}

// initGuild はBot起動時・ギルド参加時点の在室者数を現在のVoiceStatesから初期化する。
// 事前に呼んでおかないと、Bot起動時に既に通話中だったチャンネルの終了イベントで
// 誤って0人→0人（TransitionNone）と判定されてしまう
func (t *voiceCallTracker) initGuild(guildID string, voiceStates []*discordgo.VoiceState) {
	counts := make(map[string]int, len(voiceStates))
	for _, vs := range voiceStates {
		if vs.ChannelID == "" {
			continue
		}
		counts[vs.ChannelID]++
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.occupants[guildID] = counts
}

// removeGuild はBotがギルドから退出/キックされた際に該当エントリを削除する
func (t *voiceCallTracker) removeGuild(guildID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.occupants, guildID)
	delete(t.startedAt, guildID)
}

// channelTransition は1つのチャンネルについて、在室者数の増減前後からTransitionを判定し、
// 通話開始時刻を記録・回収する。increment はそのチャンネルの人数を+1するか-1するか
func (t *voiceCallTracker) channelTransition(guildID, channelID string, increment int, now time.Time) (transition voicecall.Transition, startedAt time.Time, duration time.Duration, durationKnown bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.occupants[guildID] == nil {
		t.occupants[guildID] = map[string]int{}
	}
	before := t.occupants[guildID][channelID]
	after := before + increment
	if after < 0 {
		after = 0
	}
	t.occupants[guildID][channelID] = after

	transition = voicecall.DetectTransition(before, after)
	switch transition {
	case voicecall.TransitionStarted:
		if t.startedAt[guildID] == nil {
			t.startedAt[guildID] = map[string]time.Time{}
		}
		t.startedAt[guildID][channelID] = now
		return transition, now, 0, false
	case voicecall.TransitionEnded:
		start, ok := t.startedAt[guildID][channelID]
		if ok {
			delete(t.startedAt[guildID], channelID)
			return transition, start, now.Sub(start), true
		}
		return transition, time.Time{}, 0, false
	default:
		return transition, time.Time{}, 0, false
	}
}
