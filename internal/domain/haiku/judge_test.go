package haiku

import "testing"

func TestJudge(t *testing.T) {
	tests := []struct {
		name             string
		moraCountsByWord []int
		want             bool
	}{
		{
			name:             "正常系: 単語境界が5拍・12拍で揃っていれば575と判定する",
			moraCountsByWord: []int{2, 3, 3, 4, 2, 3},
			want:             true,
		},
		{
			name:             "正常系: 古池や(5)蛙飛び込む(7)水の音(5)に相当する単語区切り",
			moraCountsByWord: []int{3, 2, 3, 4, 2, 3},
			want:             true,
		},
		{
			name:             "正常系: 単語1つが5拍・7拍・5拍ちょうどで構成されていても真",
			moraCountsByWord: []int{5, 7, 5},
			want:             true,
		},
		{
			name:             "異常系: 合計モーラ数が17拍でない場合はfalse",
			moraCountsByWord: []int{5, 7, 4},
			want:             false,
		},
		{
			name:             "異常系: 合計は17拍だが単語境界が5拍・12拍の位置に来ない場合はfalse",
			moraCountsByWord: []int{6, 6, 5},
			want:             false,
		},
		{
			name:             "異常系: 空スライスはfalse",
			moraCountsByWord: []int{},
			want:             false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Judge(tt.moraCountsByWord); got != tt.want {
				t.Errorf("Judge(%v) = %v, want %v", tt.moraCountsByWord, got, tt.want)
			}
		})
	}
}
