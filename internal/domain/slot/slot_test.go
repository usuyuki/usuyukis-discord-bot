package slot

import "testing"

func TestNewResult(t *testing.T) {
	tests := []struct {
		name  string
		reels [3]string
		want  Rank
	}{
		{name: "正常系: 3つとも一致すると大当たりになる", reels: [3]string{"a", "a", "a"}, want: RankBig},
		{name: "正常系: 先頭2つのみ一致すると小当たりになる", reels: [3]string{"a", "a", "b"}, want: RankSmall},
		{name: "正常系: 末尾2つのみ一致すると小当たりになる", reels: [3]string{"a", "b", "b"}, want: RankSmall},
		{name: "正常系: 両端のみ一致すると小当たりになる", reels: [3]string{"a", "b", "a"}, want: RankSmall},
		{name: "正常系: すべて異なるとはずれになる", reels: [3]string{"a", "b", "c"}, want: RankMiss},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewResult(tt.reels)
			if got.Rank() != tt.want {
				t.Errorf("NewResult().Rank() = %v, want %v", got.Rank(), tt.want)
			}
			if got.Reels() != tt.reels {
				t.Errorf("NewResult().Reels() = %v, want %v", got.Reels(), tt.reels)
			}
		})
	}
}
