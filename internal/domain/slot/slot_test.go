package slot

import "testing"

func TestNewResult(t *testing.T) {
	tests := []struct {
		name    string
		reels   []string
		wantErr error
		want    Rank
	}{
		{name: "正常系: 3つとも一致すると大当たりになる", reels: []string{"a", "a", "a"}, want: RankBig},
		{name: "正常系: 先頭2つのみ一致すると小当たりになる", reels: []string{"a", "a", "b"}, want: RankSmall},
		{name: "正常系: 末尾2つのみ一致すると小当たりになる", reels: []string{"a", "b", "b"}, want: RankSmall},
		{name: "正常系: 両端のみ一致すると小当たりになる", reels: []string{"a", "b", "a"}, want: RankSmall},
		{name: "正常系: すべて異なるとはずれになる", reels: []string{"a", "b", "c"}, want: RankMiss},
		{name: "異常系: reelsを2つしか入れるとErrInvalidReelCountエラーになる", reels: []string{"a", "b"}, wantErr: ErrInvalidReelCount},
		{name: "異常系: reelsを4つ入れるとErrInvalidReelCountエラーになる", reels: []string{"a", "b", "c", "d"}, wantErr: ErrInvalidReelCount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewResult(tt.reels)
			if err != tt.wantErr {
				t.Fatalf("NewResult() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if got.Rank() != tt.want {
				t.Errorf("NewResult().Rank() = %v, want %v", got.Rank(), tt.want)
			}
			for i, r := range tt.reels {
				if got.Reels()[i] != r {
					t.Errorf("NewResult().Reels()[%d] = %q, want %q", i, got.Reels()[i], r)
				}
			}
		})
	}
}
