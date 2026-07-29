package notifychannel

import "testing"

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		guildID   string
		purpose   Purpose
		channelID string
		wantErr   error
	}{
		{name: "正常系: emoji用途で全項目が埋まっていれば生成できる", guildID: "g1", purpose: PurposeEmoji, channelID: "c1", wantErr: nil},
		{name: "異常系: guildIDが空文字だとErrEmptyGuildIDになる", guildID: "", purpose: PurposeEmoji, channelID: "c1", wantErr: ErrEmptyGuildID},
		{name: "異常系: purposeが未定義値だとErrInvalidPurposeになる", guildID: "g1", purpose: Purpose("unknown"), channelID: "c1", wantErr: ErrInvalidPurpose},
		{name: "異常系: channelIDが空文字だとErrEmptyChannelIDになる", guildID: "g1", purpose: PurposeEmoji, channelID: "", wantErr: ErrEmptyChannelID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.guildID, tt.purpose, tt.channelID)
			if err != tt.wantErr {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestPurpose_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		purpose Purpose
		want    bool
	}{
		{name: "正常系: emojiは有効な値", purpose: PurposeEmoji, want: true},
		{name: "異常系: 未定義の値は無効", purpose: Purpose("unknown"), want: false},
		{name: "異常系: 空文字は無効", purpose: Purpose(""), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.purpose.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
