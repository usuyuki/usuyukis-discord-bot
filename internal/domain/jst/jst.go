// Package jst は日本標準時（UTC+9）の時刻整形に関する共有ロジックを提供する
package jst

import "time"

// Location は日本標準時（UTC+9）。実行環境にtzdataが無くても解決できるようFixedZoneで定義する
var Location = time.FixedZone("Asia/Tokyo", 9*60*60)

// layout はFormatで使う日時表示フォーマット
const layout = "2006-01-02 15:04:05"

// Format はtをJSTの日時文字列（YYYY-MM-DD HH:MM:SS）に整形する
func Format(t time.Time) string {
	return t.In(Location).Format(layout)
}
