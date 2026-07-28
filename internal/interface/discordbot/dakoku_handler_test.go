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
			name:        "正常系: Botへのメンションがあれば現在時刻を返信する",
			msg:         IncomingMessage{ChannelID: "c1", MentionsBotID: true},
			wantCalled:  true,
			wantContent: "2026-07-28 14:32:10",
		},
		{
			name:       "異常系: Botへのメンションがなければ何もしない",
			msg:        IncomingMessage{ChannelID: "c1", MentionsBotID: false},
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
