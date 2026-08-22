package discord

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{name: "正常系: 1時間29分4秒はHH:MM:SS形式に整形される", d: 1*time.Hour + 29*time.Minute + 4*time.Second, want: "01:29:04"},
		{name: "正常系: 0秒は00:00:00になる", d: 0, want: "00:00:00"},
		{name: "正常系: 1時間未満は時が0埋めの2桁になる", d: 5 * time.Minute, want: "00:05:00"},
		{name: "正常系: 100時間を超えても時の桁は2桁未満に切り詰められず3桁以上で表示される", d: 100 * time.Hour, want: "100:00:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.d); got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestFormatJST(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{
			name: "正常系: UTC時刻はJST（+9時間）に変換されて整形される",
			t:    time.Date(2026, 6, 8, 16, 52, 51, 0, time.UTC),
			want: "2026-06-09 01:52:51",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatJST(tt.t); got != tt.want {
				t.Errorf("formatJST(%v) = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}
