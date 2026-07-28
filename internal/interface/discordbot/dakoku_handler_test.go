package discordbot

import (
	"context"
	"testing"
	"time"
)

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

func TestDakokuHandler_HandleMessage(t *testing.T) {
	fixedNow := time.Date(2026, 7, 28, 14, 32, 10, 0, time.UTC)

	tests := []struct {
		name        string
		msg         IncomingMessage
		wantCalled  bool
		wantContent string
	}{
		{
			name:        "正常系: 本文に@Rakuroが含まれれば現在時刻を返信する",
			msg:         IncomingMessage{ChannelID: "c1", Content: "@Rakuro 今何時？"},
			wantCalled:  true,
			wantContent: "2026-07-28 14:32:10",
		},
		{
			name:        "正常系: 本文に@rakuro（小文字）が含まれれば現在時刻を返信する",
			msg:         IncomingMessage{ChannelID: "c1", Content: "@rakuro 今何時？"},
			wantCalled:  true,
			wantContent: "2026-07-28 14:32:10",
		},
		{
			name:       "異常系: Botへのメンションが一切含まれなければ何もしない",
			msg:        IncomingMessage{ChannelID: "c1", Content: "こんにちは"},
			wantCalled: false,
		},
		{
			name:       "異常系: keywordコマンドであれば打刻は反応しない",
			msg:        IncomingMessage{ChannelID: "c1", Content: "@Rakuro keyword add ぬるぽ ガッ"},
			wantCalled: false,
		},
		{
			name:       "異常系: 無関係な単語に@rakuroが連結しているだけなら反応しない",
			msg:        IncomingMessage{ChannelID: "c1", Content: "check https://example.com/@rakuromusic"},
			wantCalled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeMessageSender{}
			h := NewDakokuHandler(sender)
			h.now = func() time.Time { return fixedNow }

			if err := h.HandleMessage(context.Background(), tt.msg); err != nil {
				t.Fatalf("HandleMessage() unexpected error = %v", err)
			}
			if sender.called != tt.wantCalled {
				t.Fatalf("HandleMessage() called = %v, want %v", sender.called, tt.wantCalled)
			}
			if tt.wantCalled && sender.sentContent != tt.wantContent {
				t.Errorf("HandleMessage() content = %q, want %q", sender.sentContent, tt.wantContent)
			}
		})
	}
}
