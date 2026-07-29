package dakoku

import "time"

// timeLayout は打刻時刻の表示フォーマット
const timeLayout = "2006-01-02 15:04:05"

// jst は日本標準時（UTC+9）。実行環境にtzdataが無くても解決できるようFixedZoneで定義する
var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// FormatNow は現在時刻をJSTに変換し、打刻表示用フォーマットの文字列に変換する
func FormatNow(now time.Time) string {
	return now.In(jst).Format(timeLayout)
}
