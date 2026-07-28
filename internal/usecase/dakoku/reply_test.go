package dakoku

import (
	"testing"
	"time"
)

func TestReply(t *testing.T) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{
			name: "正常系: 指定時刻をフォーマットした文字列を返す",
			now:  time.Date(2026, 7, 28, 14, 32, 10, 0, jst),
			want: "2026-07-28 14:32:10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reply(tt.now); got != tt.want {
				t.Errorf("Reply() = %q, want %q", got, tt.want)
			}
		})
	}
}
