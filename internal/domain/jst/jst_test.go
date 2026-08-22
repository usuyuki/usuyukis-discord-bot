package jst

import (
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
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
			if got := Format(tt.t); got != tt.want {
				t.Errorf("Format(%v) = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}
