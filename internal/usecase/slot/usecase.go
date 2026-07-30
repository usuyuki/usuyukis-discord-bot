package slot

import (
	"context"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/slot"
)

// reelCount はスロットのリール数（抽選する絵文字の個数）
const reelCount = 3

// fallbackEmojis はギルドのカスタム絵文字がreelCount未満のときに使う標準絵文字セット。
// カスタム絵文字が少ない/未登録のギルドでもslotを遊べるようにするためのフォールバック
var fallbackEmojis = []string{"🍒", "🍋", "🔔", "💎", "7️⃣", "🍇", "🍉", "⭐"}

// UseCase はスロット抽選に関するアプリケーションロジック
type UseCase struct {
	emojiSource EmojiSource
	// intn はrandomな整数を1つ生成する関数。テスト容易性のため注入可能にしている
	intn func(n int) int
}

// New はUseCaseを生成する
func New(emojiSource EmojiSource, intn func(n int) int) *UseCase {
	return &UseCase{emojiSource: emojiSource, intn: intn}
}

// Spin はギルドのカスタム絵文字（reelCount個未満ならfallbackEmojis）からreelCount個を
// 独立にランダム抽選し、役判定込みのResultを返す
func (u *UseCase) Spin(ctx context.Context, guildID string) (slot.Result, error) {
	source, err := u.emojiSource.ListEmojiTags(ctx, guildID)
	if err != nil {
		return slot.Result{}, err
	}
	if len(source) < reelCount {
		source = fallbackEmojis
	}

	reels := make([]string, reelCount)
	for i := range reels {
		reels[i] = source[u.intn(len(source))]
	}
	return slot.NewResult(reels)
}
