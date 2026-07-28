package dakoku

import "time"

// timeLayout は打刻時刻の表示フォーマット
const timeLayout = "2006-01-02 15:04:05"

// FormatNow は現在時刻を打刻表示用フォーマットの文字列に変換する
func FormatNow(now time.Time) string {
	return now.Format(timeLayout)
}
