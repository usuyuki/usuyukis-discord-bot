package dakoku

import (
	"time"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/dakoku"
)

// Reply は現在時刻を打刻表示用文字列にして返すユースケース
func Reply(now time.Time) string {
	return dakoku.FormatNow(now)
}
