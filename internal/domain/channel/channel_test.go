package channel

import (
	"strings"
	"testing"
)

func TestNewName(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
		want    string
	}{
		{
			name: "正常系: 英数字とハイフンのみの名前はそのまま登録できる",
			raw:  "general-chat2",
			want: "general-chat2",
		},
		{
			name: "正常系: アンダースコアを含む名前も登録できる",
			raw:  "my_channel",
			want: "my_channel",
		},
		{
			name: "正常系: ひらがな・カタカナ・漢字を含む名前も登録できる",
			raw:  "雑談チャンネルその1",
			want: "雑談チャンネルその1",
		},
		{
			name: "正常系: 日本語とハイフン・アンダースコアの混在も登録できる",
			raw:  "お知らせ-general_ch",
			want: "お知らせ-general_ch",
		},
		{
			name:    "異常系: 空文字を入れると名前が空になるのでエラーになる",
			raw:     "",
			wantErr: true,
		},
		{
			name:    "異常系: 大文字を入れるとDiscordの命名規則に反するのでエラーになる",
			raw:     "General",
			wantErr: true,
		},
		{
			name:    "異常系: スペースを入れると許容文字外なのでエラーになる",
			raw:     "my channel",
			wantErr: true,
		},
		{
			name:    "異常系: 絵文字を入れると許容文字外なのでエラーになる",
			raw:     "channel🎉",
			wantErr: true,
		},
		{
			name:    "異常系: 101文字を入れると最大長を超えるのでエラーになる",
			raw:     strings.Repeat("a", 101),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewName(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewName(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.String() != tt.want {
				t.Errorf("NewName(%q).String() = %q, want %q", tt.raw, got.String(), tt.want)
			}
		})
	}
}
