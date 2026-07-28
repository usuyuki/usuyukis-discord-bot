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
		{
			name:             "異常系: 短歌相当(31拍)はJudge単体ではfalse",
			moraCountsByWord: []int{5, 7, 5, 7, 7},
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

func TestJudgeTanka(t *testing.T) {
	tests := []struct {
		name             string
		moraCountsByWord []int
		want             bool
	}{
		{
			name:             "正常系: 単語境界が5拍・12拍・17拍・24拍で揃っていれば57577と判定する",
			moraCountsByWord: []int{2, 3, 3, 4, 2, 3, 3, 4, 4, 3},
			want:             true,
		},
		{
			name:             "正常系: 単語1つが5,7,5,7,7拍ちょうどで構成されていても真",
			moraCountsByWord: []int{5, 7, 5, 7, 7},
			want:             true,
		},
		{
			name:             "異常系: 合計モーラ数が31拍でない場合はfalse",
			moraCountsByWord: []int{5, 7, 5, 7, 6},
			want:             false,
		},
		{
			name:             "異常系: 合計は31拍だが単語境界が区切り位置に来ない場合はfalse",
			moraCountsByWord: []int{6, 6, 6, 6, 7},
			want:             false,
		},
		{
			name:             "異常系: 俳句相当(17拍)はJudgeTanka単体ではfalse",
			moraCountsByWord: []int{5, 7, 5},
			want:             false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JudgeTanka(tt.moraCountsByWord); got != tt.want {
				t.Errorf("JudgeTanka(%v) = %v, want %v", tt.moraCountsByWord, got, tt.want)
			}
		})
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name    string
		words   []Word
		pattern []int
		want    []string
		wantOK  bool
	}{
		{
			name: "正常系: 575パターンで単語境界が一致すれば区切り文字列を返す",
			words: []Word{
				{Surface: "古池", MoraCount: 3},
				{Surface: "や", MoraCount: 2},
				{Surface: "蛙", MoraCount: 3},
				{Surface: "飛び込む", MoraCount: 4},
				{Surface: "水の", MoraCount: 3},
				{Surface: "音", MoraCount: 2},
			},
			pattern: []int{5, 7, 5},
			want:    []string{"古池や", "蛙飛び込む", "水の音"},
			wantOK:  true,
		},
		{
			name: "異常系: 単語境界が区切り位置に一致しなければfalseと空スライスを返す",
			words: []Word{
				{Surface: "ころころ", MoraCount: 6},
				{Surface: "ころころ", MoraCount: 6},
				{Surface: "ころ", MoraCount: 5},
			},
			pattern: []int{5, 7, 5},
			want:    nil,
			wantOK:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Split(tt.words, tt.pattern)
			if ok != tt.wantOK {
				t.Fatalf("Split() ok = %v, want %v", ok, tt.wantOK)
			}
			if !equalStringSlice(got, tt.want) {
				t.Errorf("Split() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
