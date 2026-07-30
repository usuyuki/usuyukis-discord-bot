package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestConvertEmojis(t *testing.T) {
	tests := []struct {
		name string
		src  []*discordgo.Emoji
		want []string
	}{
		{
			name: "正常系: name/idが揃った絵文字はそのままTagに変換される",
			src: []*discordgo.Emoji{
				{ID: "1", Name: "a", Animated: false},
				{ID: "2", Name: "b", Animated: true},
			},
			want: []string{"<:a:1>", "<a:b:2>"},
		},
		{
			name: "異常系: nameが空の絵文字はemoji.Newが失敗するため結果から除外される",
			src: []*discordgo.Emoji{
				{ID: "1", Name: ""},
				{ID: "2", Name: "b"},
			},
			want: []string{"<:b:2>"},
		},
		{
			name: "異常系: idが空の絵文字はemoji.Newが失敗するため結果から除外される",
			src: []*discordgo.Emoji{
				{ID: "", Name: "a"},
				{ID: "2", Name: "b"},
			},
			want: []string{"<:b:2>"},
		},
		{
			name: "正常系: 空スライスを渡すと空スライスを返す",
			src:  []*discordgo.Emoji{},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertEmojis(tt.src)
			gotTags := make([]string, len(got))
			for i, de := range got {
				gotTags[i] = de.Tag()
			}
			if !equalStringSlice(gotTags, tt.want) {
				t.Errorf("convertEmojis() tags = %v, want %v", gotTags, tt.want)
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
