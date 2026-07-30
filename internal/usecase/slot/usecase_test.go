package slot

import (
	"context"
	"errors"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/slot"
)

type fakeEmojiSource struct {
	tags []string
	err  error
}

func (f *fakeEmojiSource) ListEmojiTags(ctx context.Context, guildID string) ([]string, error) {
	return f.tags, f.err
}

// fakeRandomizer はあらかじめ用意したindexesを呼び出し順に返すテスト用Randomizer。
// 呼び出し回数がindexesの長さを超えたらパニックする
type fakeRandomizer struct {
	indexes []int
	call    int
}

func (f *fakeRandomizer) Intn(n int) int {
	i := f.indexes[f.call]
	f.call++
	return i
}

func TestUseCase_Spin(t *testing.T) {
	tests := []struct {
		name      string
		guildID   string
		tags      []string
		sourceErr error
		indexes   []int
		wantReels [3]string
		wantRank  slot.Rank
		wantErr   bool
	}{
		{
			name:      "正常系: カスタム絵文字が3つ以上あればそこから抽選する",
			guildID:   "g1",
			tags:      []string{"<:a:1>", "<:b:2>", "<:c:3>"},
			indexes:   []int{0, 0, 0},
			wantReels: [3]string{"<:a:1>", "<:a:1>", "<:a:1>"},
			wantRank:  slot.RankBig,
		},
		{
			name:      "正常系: カスタム絵文字がreelCount未満ならfallbackEmojisから抽選する",
			guildID:   "g1",
			tags:      []string{"<:a:1>"},
			indexes:   []int{0, 1, 2},
			wantReels: [3]string{fallbackEmojis[0], fallbackEmojis[1], fallbackEmojis[2]},
			wantRank:  slot.RankMiss,
		},
		{
			name:      "正常系: カスタム絵文字が0個でもfallbackEmojisから抽選する",
			guildID:   "g1",
			tags:      nil,
			indexes:   []int{0, 0, 1},
			wantReels: [3]string{fallbackEmojis[0], fallbackEmojis[0], fallbackEmojis[1]},
			wantRank:  slot.RankSmall,
		},
		{
			name:      "正常系: guildIDが空文字（DM）ならEmojiSourceを呼ばずfallbackEmojisから抽選する",
			guildID:   "",
			sourceErr: errors.New("boom"),
			indexes:   []int{0, 0, 1},
			wantReels: [3]string{fallbackEmojis[0], fallbackEmojis[0], fallbackEmojis[1]},
			wantRank:  slot.RankSmall,
		},
		{
			name:      "異常系: EmojiSourceがエラーを返すとSpinもエラーを返す",
			guildID:   "g1",
			sourceErr: errors.New("boom"),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := &fakeEmojiSource{tags: tt.tags, err: tt.sourceErr}
			u := New(source, &fakeRandomizer{indexes: tt.indexes})

			got, err := u.Spin(context.Background(), tt.guildID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Spin() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Spin() unexpected error = %v", err)
			}
			if got.Reels() != tt.wantReels {
				t.Errorf("Spin().Reels() = %v, want %v", got.Reels(), tt.wantReels)
			}
			if got.Rank() != tt.wantRank {
				t.Errorf("Spin().Rank() = %v, want %v", got.Rank(), tt.wantRank)
			}
		})
	}
}
