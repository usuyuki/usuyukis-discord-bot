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
	randomizer  Randomizer
}

// New はUseCaseを生成する
func New(emojiSource EmojiSource, randomizer Randomizer) *UseCase {
	return &UseCase{emojiSource: emojiSource, randomizer: randomizer}
}

// Spin はギルドのカスタム絵文字（reelCount個未満ならfallbackEmojis）からreelCount個を
// 独立にランダム抽選し、役判定込みのResultを返す。
// guildIDが空文字（DM等ギルドに属さないメッセージ）の場合はカスタム絵文字を取得できないため、
// EmojiSourceを呼ばずfallbackEmojisを使う
func (u *UseCase) Spin(ctx context.Context, guildID string) (slot.Result, error) {
	source := fallbackEmojis
	if guildID != "" {
		tags, err := u.emojiSource.ListEmojiTags(ctx, guildID)
		if err != nil {
			return slot.Result{}, err
		}
		if len(tags) >= reelCount {
			source = tags
		}
	}

	var reels [reelCount]string
	for i := range reels {
		reels[i] = source[u.randomizer.Intn(len(source))]
	}
	return slot.NewResult(reels), nil
}
