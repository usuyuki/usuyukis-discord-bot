package dakoku

import (
	"testing"
	"time"
)

func TestFormatNow(t *testing.T) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{
			name: "正常系: 指定時刻をyyyy-MM-dd HH:mm:ss形式の文字列に変換する",
			now:  time.Date(2026, 7, 28, 14, 32, 10, 0, jst),
			want: "2026-07-28 14:32:10",
		},
		{
			name: "正常系: 1桁の月日時分秒もゼロ埋めされる",
			now:  time.Date(2026, 1, 2, 3, 4, 5, 0, jst),
			want: "2026-01-02 03:04:05",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatNow(tt.now); got != tt.want {
				t.Errorf("FormatNow() = %q, want %q", got, tt.want)
			}
		})
	}
}
