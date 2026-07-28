package discordbot

import "testing"

func TestIncomingMessage_MentionsRakuro(t *testing.T) {
	tests := []struct {
		name string
		msg  IncomingMessage
		want bool
	}{
		{
			name: "正常系: 構造化メンション（MentionsBotID）があれば真",
			msg:  IncomingMessage{MentionsBotID: true, Content: "こんにちは"},
			want: true,
		},
		{
			name: "正常系: 本文にテキストの@Rakuro表記が含まれれば真",
			msg:  IncomingMessage{MentionsBotID: false, Content: "@Rakuro 今何時？"},
			want: true,
		},
		{
			name: "異常系: 構造化メンションもテキストの@Rakuro表記もなければ偽",
			msg:  IncomingMessage{MentionsBotID: false, Content: "こんにちは"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.MentionsRakuro(); got != tt.want {
				t.Errorf("MentionsRakuro() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMentionsRakuroText(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "正常系: 大文字の@Rakuroを検知する",
			content: "@Rakuro 今何時？",
			want:    true,
		},
		{
			name:    "正常系: 小文字の@rakuroを検知する",
			content: "@rakuro 今何時？",
			want:    true,
		},
		{
			name:    "正常系: 句読点が後続していても単語境界として検知する",
			content: "@Rakuro、今何時？",
			want:    true,
		},
		{
			name:    "異常系: @rakuroという文字列が含まれなければ偽",
			content: "こんにちは",
			want:    false,
		},
		{
			name:    "異常系: URLの一部など別の単語に連結している場合は偽",
			content: "check https://example.com/@rakuromusic",
			want:    false,
		},
		{
			name:    "異常系: 別アカウント名の接頭辞に一致するだけの場合は偽",
			content: "@Rakuroski さんによろしく",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mentionsRakuroText(tt.content); got != tt.want {
				t.Errorf("mentionsRakuroText(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}
