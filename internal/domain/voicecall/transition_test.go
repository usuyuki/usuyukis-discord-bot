package voicecall

import "testing"

func TestDetectTransition(t *testing.T) {
	tests := []struct {
		name        string
		beforeCount int
		afterCount  int
		want        Transition
	}{
		{name: "正常系: 0人から1人になると通話開始と判定する", beforeCount: 0, afterCount: 1, want: TransitionStarted},
		{name: "正常系: 0人から2人になっても通話開始と判定する", beforeCount: 0, afterCount: 2, want: TransitionStarted},
		{name: "正常系: 1人から0人になると通話終了と判定する", beforeCount: 1, afterCount: 0, want: TransitionEnded},
		{name: "正常系: 3人から0人になっても通話終了と判定する", beforeCount: 3, afterCount: 0, want: TransitionEnded},
		{name: "異常系: 1人から2人に増えても0人をまたがないので変化なしと判定する", beforeCount: 1, afterCount: 2, want: TransitionNone},
		{name: "異常系: 2人から1人に減っても0人をまたがないので変化なしと判定する", beforeCount: 2, afterCount: 1, want: TransitionNone},
		{name: "異常系: 0人のまま変化がなければ変化なしと判定する", beforeCount: 0, afterCount: 0, want: TransitionNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectTransition(tt.beforeCount, tt.afterCount); got != tt.want {
				t.Errorf("DetectTransition(%d, %d) = %v, want %v", tt.beforeCount, tt.afterCount, got, tt.want)
			}
		})
	}
}
