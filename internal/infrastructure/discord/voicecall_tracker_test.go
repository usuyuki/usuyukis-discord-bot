package discord

import (
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/voicecall"
)

var t0 = time.Date(2026, 6, 9, 1, 52, 51, 0, time.UTC)

func TestVoiceCallTracker_ChannelTransition(t *testing.T) {
	t.Run("正常系: 誰もいないチャンネルに1人入ると開始と判定され開始時刻が記録される", func(t *testing.T) {
		tr := newVoiceCallTracker()

		transition, startedAt, _, durationKnown := tr.channelTransition("g1", "c1", 1, t0)

		if transition != voicecall.TransitionStarted {
			t.Fatalf("transition = %v, want %v", transition, voicecall.TransitionStarted)
		}
		if !startedAt.Equal(t0) {
			t.Errorf("startedAt = %v, want %v", startedAt, t0)
		}
		if durationKnown {
			t.Errorf("durationKnown = true, want false for started transition")
		}
	})

	t.Run("正常系: 1人だけのチャンネルから退室すると終了と判定され開始からの経過時間が返る", func(t *testing.T) {
		tr := newVoiceCallTracker()
		tr.channelTransition("g1", "c1", 1, t0)

		endAt := t0.Add(89 * time.Minute)
		transition, _, duration, durationKnown := tr.channelTransition("g1", "c1", -1, endAt)

		if transition != voicecall.TransitionEnded {
			t.Fatalf("transition = %v, want %v", transition, voicecall.TransitionEnded)
		}
		if !durationKnown {
			t.Fatal("durationKnown = false, want true")
		}
		if duration != 89*time.Minute {
			t.Errorf("duration = %v, want %v", duration, 89*time.Minute)
		}
	})

	t.Run("正常系: 2人目が入っても在室者は0人をまたがないので変化なしと判定される", func(t *testing.T) {
		tr := newVoiceCallTracker()
		tr.channelTransition("g1", "c1", 1, t0)

		transition, _, _, _ := tr.channelTransition("g1", "c1", 1, t0.Add(time.Minute))

		if transition != voicecall.TransitionNone {
			t.Errorf("transition = %v, want %v", transition, voicecall.TransitionNone)
		}
	})

	t.Run("正常系: 3人中1人退室しても0人をまたがないので変化なしと判定される", func(t *testing.T) {
		tr := newVoiceCallTracker()
		tr.channelTransition("g1", "c1", 1, t0)
		tr.channelTransition("g1", "c1", 1, t0)
		tr.channelTransition("g1", "c1", 1, t0)

		transition, _, _, _ := tr.channelTransition("g1", "c1", -1, t0.Add(time.Minute))

		if transition != voicecall.TransitionNone {
			t.Errorf("transition = %v, want %v", transition, voicecall.TransitionNone)
		}
	})

	t.Run("異常系: startedAt記録前にchannelTransitionでoccupantsだけが増えた状態から終了イベントが来るとdurationKnownはfalseになる", func(t *testing.T) {
		tr := newVoiceCallTracker()
		// occupantsだけを直接操作し、startedAtが記録されていない状況を再現する
		// （通常のイベントフローでは発生しないが、防御的に確認しておく）
		tr.mu.Lock()
		tr.occupants["g1"] = map[string]int{"c1": 1}
		tr.mu.Unlock()

		transition, _, _, durationKnown := tr.channelTransition("g1", "c1", -1, t0)

		if transition != voicecall.TransitionEnded {
			t.Fatalf("transition = %v, want %v", transition, voicecall.TransitionEnded)
		}
		if durationKnown {
			t.Errorf("durationKnown = true, want false when start time was never recorded")
		}
	})

	t.Run("異常系: 在室者0人のチャンネルから退室イベントが来ても人数が負にならず変化なしと判定される", func(t *testing.T) {
		tr := newVoiceCallTracker()

		transition, _, _, _ := tr.channelTransition("g1", "c1", -1, t0)

		if transition != voicecall.TransitionNone {
			t.Errorf("transition = %v, want %v", transition, voicecall.TransitionNone)
		}
	})
}

func TestVoiceCallTracker_InitGuild(t *testing.T) {
	t.Run("正常系: 起動時に既に在室者がいるチャンネルは初期化され、その後の退室で終了と判定される", func(t *testing.T) {
		tr := newVoiceCallTracker()
		tr.initGuild("g1", []*discordgo.VoiceState{
			{ChannelID: "c1", UserID: "u1"},
			{ChannelID: "c1", UserID: "u2"},
			{ChannelID: "c2", UserID: "u3"},
		}, time.Now())

		// c1は2人在室中なので1人減っても終了にならない
		transition, _, _, _ := tr.channelTransition("g1", "c1", -1, time.Now())
		if transition != voicecall.TransitionNone {
			t.Errorf("transition = %v, want %v", transition, voicecall.TransitionNone)
		}

		// c2は1人在室中なので退室で終了と判定される
		transition, _, _, _ = tr.channelTransition("g1", "c2", -1, time.Now())
		if transition != voicecall.TransitionEnded {
			t.Errorf("transition = %v, want %v", transition, voicecall.TransitionEnded)
		}
	})

	t.Run("正常系: 初期化されたチャンネルはstartedAtも記録されており、終了時にdurationKnownがtrueになる", func(t *testing.T) {
		tr := newVoiceCallTracker()
		initAt := t0
		tr.initGuild("g1", []*discordgo.VoiceState{{ChannelID: "c1", UserID: "u1"}}, initAt)

		endAt := initAt.Add(10 * time.Minute)
		transition, startedAt, duration, durationKnown := tr.channelTransition("g1", "c1", -1, endAt)

		if transition != voicecall.TransitionEnded {
			t.Fatalf("transition = %v, want %v", transition, voicecall.TransitionEnded)
		}
		if !durationKnown {
			t.Fatal("durationKnown = false, want true")
		}
		if !startedAt.Equal(initAt) {
			t.Errorf("startedAt = %v, want %v", startedAt, initAt)
		}
		if duration != 10*time.Minute {
			t.Errorf("duration = %v, want %v", duration, 10*time.Minute)
		}
	})

	t.Run("正常系: 既にstartedAtが記録済みのチャンネルはGuildCreate再送のinitGuildで上書きされない", func(t *testing.T) {
		tr := newVoiceCallTracker()
		tr.channelTransition("g1", "c1", 1, t0)

		// GuildCreate再送を模して、より新しい時刻でinitGuildを呼ぶ
		tr.initGuild("g1", []*discordgo.VoiceState{{ChannelID: "c1", UserID: "u1"}}, t0.Add(time.Hour))

		endAt := t0.Add(90 * time.Minute)
		_, startedAt, duration, durationKnown := tr.channelTransition("g1", "c1", -1, endAt)

		if !durationKnown {
			t.Fatal("durationKnown = false, want true")
		}
		if !startedAt.Equal(t0) {
			t.Errorf("startedAt = %v, want %v (should not be overwritten by re-init)", startedAt, t0)
		}
		if duration != 90*time.Minute {
			t.Errorf("duration = %v, want %v", duration, 90*time.Minute)
		}
	})
}

func TestVoiceCallTracker_WithGuildLock(t *testing.T) {
	t.Run("正常系: 同一ギルドに対する複数呼び出しは直列に実行される", func(t *testing.T) {
		tr := newVoiceCallTracker()
		var (
			mu      sync.Mutex
			running bool
			overlap bool
			done    sync.WaitGroup
		)

		for range 20 {
			done.Add(1)
			go func() {
				defer done.Done()
				tr.withGuildLock("g1", func() {
					mu.Lock()
					if running {
						overlap = true
					}
					running = true
					mu.Unlock()

					time.Sleep(time.Millisecond)

					mu.Lock()
					running = false
					mu.Unlock()
				})
			}()
		}
		done.Wait()

		if overlap {
			t.Error("withGuildLock allowed overlapping execution for the same guild")
		}
	})

	t.Run("正常系: 異なるギルドの呼び出しはロックを共有しない", func(t *testing.T) {
		tr := newVoiceCallTracker()
		tr.withGuildLock("g1", func() {})
		tr.withGuildLock("g2", func() {})

		tr.guildMu.Lock()
		_, g1ok := tr.guilds["g1"]
		_, g2ok := tr.guilds["g2"]
		tr.guildMu.Unlock()

		if !g1ok || !g2ok {
			t.Errorf("expected locks for both g1 and g2 to be registered, got g1=%v g2=%v", g1ok, g2ok)
		}
	})
}

func TestVoiceCallTracker_RemoveGuild(t *testing.T) {
	t.Run("正常系: ギルド削除後は在室者数がリセットされ次の入室が開始と判定される", func(t *testing.T) {
		tr := newVoiceCallTracker()
		tr.channelTransition("g1", "c1", 1, time.Now())

		tr.removeGuild("g1")

		transition, _, _, _ := tr.channelTransition("g1", "c1", 1, time.Now())
		if transition != voicecall.TransitionStarted {
			t.Errorf("transition = %v, want %v", transition, voicecall.TransitionStarted)
		}
	})

	t.Run("正常系: ギルド削除後はwithGuildLockのロックエントリも削除される", func(t *testing.T) {
		tr := newVoiceCallTracker()
		tr.withGuildLock("g1", func() {})

		tr.removeGuild("g1")

		tr.guildMu.Lock()
		_, ok := tr.guilds["g1"]
		tr.guildMu.Unlock()
		if ok {
			t.Error("expected guild lock entry to be removed after removeGuild")
		}
	})
}
