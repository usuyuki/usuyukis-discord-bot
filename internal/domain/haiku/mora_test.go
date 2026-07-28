package haiku

import (
	"reflect"
	"testing"
)

func TestSplitMorae(t *testing.T) {
	tests := []struct {
		name    string
		reading string
		want    []string
	}{
		{
			name:    "正常系: 促音・撥音・長音は独立した1モーラになる",
			reading: "ガッコー",
			want:    []string{"ガ", "ッ", "コ", "ー"},
		},
		{
			name:    "正常系: 拗音は直前の文字と結合して1モーラになる",
			reading: "キャベツ",
			want:    []string{"キャ", "ベ", "ツ"},
		},
		{
			name:    "正常系: 通常のひらがなはそのまま1文字1モーラになる",
			reading: "はいく",
			want:    []string{"は", "い", "く"},
		},
		{
			name:    "正常系: 空文字は空スライスになる",
			reading: "",
			want:    []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitMorae(tt.reading)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitMorae() = %v, want %v", got, tt.want)
			}
		})
	}
}
