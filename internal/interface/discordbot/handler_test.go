package discordbot

import (
	"context"
	"reflect"
	"testing"
)

// fakeMessageSender はテスト用のMessageSenderフェイク実装。
// 各ハンドラのテストから共通で利用する
type fakeMessageSender struct {
	called        bool
	sentChannelID string
	sentContent   string
}

func (f *fakeMessageSender) SendMessage(ctx context.Context, channelID, content string) error {
	f.called = true
	f.sentChannelID = channelID
	f.sentContent = content
	return nil
}

func TestMentionTag(t *testing.T) {
	tests := []struct {
		name  string
		botID string
		want  string
	}{
		{
			name:  "正常系: botIDから構造化メンション表記を組み立てる",
			botID: "123456789",
			want:  "<@123456789>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mentionTag(tt.botID); got != tt.want {
				t.Errorf("mentionTag(%q) = %q, want %q", tt.botID, got, tt.want)
			}
		})
	}
}

func TestStripMentionTokens(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		want   []string
	}{
		{
			name:   "正常系: 構造化メンショントークンのみ除去し残りのフィールドを保持する",
			fields: []string{"<@123>", "keyword", "add"},
			want:   []string{"keyword", "add"},
		},
		{
			name:   "正常系: 複数の構造化メンションが混ざっていてもすべて除去する",
			fields: []string{"<@123>", "<@&456>", "help"},
			want:   []string{"help"},
		},
		{
			name:   "異常系: メンショントークンがなければそのまま返す",
			fields: []string{"help"},
			want:   []string{"help"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripMentionTokens(tt.fields)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stripMentionTokens(%v) = %v, want %v", tt.fields, got, tt.want)
			}
		})
	}
}
