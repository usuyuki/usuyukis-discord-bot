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
		name    string
		fields  []string
		botID   string
		botName string
		want    []string
	}{
		{
			name:   "正常系: BotIDの構造化メンショントークンのみ除去し残りのフィールドを保持する",
			fields: []string{"<@123>", "keyword", "add"},
			botID:  "123",
			want:   []string{"keyword", "add"},
		},
		{
			name:   "正常系: BotIDのメンションが複数箇所にあってもすべて除去する",
			fields: []string{"<@123>", "help", "<@123>"},
			botID:  "123",
			want:   []string{"help"},
		},
		{
			name:   "異常系: メンショントークンがなければそのまま返す",
			fields: []string{"help"},
			botID:  "123",
			want:   []string{"help"},
		},
		{
			name:   "異常系: BotID以外のメンション形式の文字列はキーワード引数として保持し除去しない",
			fields: []string{"<@123>", "keyword", "add", "<@notanid>", "response"},
			botID:  "123",
			want:   []string{"keyword", "add", "<@notanid>", "response"},
		},
		{
			name:    "正常系: 先頭フィールドが@付きbotNameと一致すれば除去する（コピペ救済）",
			fields:  []string{"@usuyuki", "help"},
			botID:   "123",
			botName: "usuyuki",
			want:    []string{"help"},
		},
		{
			name:    "正常系: 先頭フィールドのbotName一致は大文字小文字を無視する",
			fields:  []string{"@Usuyuki", "help"},
			botID:   "123",
			botName: "usuyuki",
			want:    []string{"help"},
		},
		{
			name:    "異常系: botNameと一致しても先頭以外のフィールドは除去しない",
			fields:  []string{"keyword", "add", "@usuyuki", "response"},
			botID:   "123",
			botName: "usuyuki",
			want:    []string{"keyword", "add", "@usuyuki", "response"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripMentionTokens(tt.fields, tt.botID, tt.botName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stripMentionTokens(%v, %q, %q) = %v, want %v", tt.fields, tt.botID, tt.botName, got, tt.want)
			}
		})
	}
}
