package discordbot

import (
	"context"
	"testing"

	slotUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/slot"
)

type fakeSlotEmojiSource struct {
	tags []string
}

func (f *fakeSlotEmojiSource) ListEmojiTags(ctx context.Context, guildID string) ([]string, error) {
	return f.tags, nil
}

func TestSlotHandler_HandleMessage(t *testing.T) {
	tests := []struct {
		name        string
		msg         IncomingMessage
		wantSent    bool
		wantContent string
	}{
		{
			name:        "正常系: @Bot slot で3つ揃えば大当たりを通知する",
			msg:         IncomingMessage{GuildID: "g1", ChannelID: "c1", Content: "<@bot123> slot", MentionsBotID: true, BotID: "bot123"},
			wantSent:    true,
			wantContent: "<:a:1> | <:a:1> | <:a:1>\n🎉 大当たり！",
		},
		{
			name:     "異常系: Botへのメンションでなければ反応しない",
			msg:      IncomingMessage{GuildID: "g1", ChannelID: "c1", Content: "slot", MentionsBotID: false, BotID: "bot123"},
			wantSent: false,
		},
		{
			name:     "異常系: slot以外のコマンドには反応しない",
			msg:      IncomingMessage{GuildID: "g1", ChannelID: "c1", Content: "<@bot123> keyword", MentionsBotID: true, BotID: "bot123"},
			wantSent: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeMessageSender{}
			uc := slotUC.New(&fakeSlotEmojiSource{tags: []string{"<:a:1>", "<:b:2>", "<:c:3>"}}, func(n int) int { return 0 })
			h := NewSlotHandler(uc, sender)

			if err := h.HandleMessage(context.Background(), tt.msg); err != nil {
				t.Fatalf("HandleMessage() unexpected error = %v", err)
			}
			if sender.called != tt.wantSent {
				t.Fatalf("HandleMessage() sent = %v, want %v", sender.called, tt.wantSent)
			}
			if tt.wantSent {
				if sender.sentChannelID != tt.msg.ChannelID {
					t.Errorf("HandleMessage() sentChannelID = %q, want %q", sender.sentChannelID, tt.msg.ChannelID)
				}
				if sender.sentContent != tt.wantContent {
					t.Errorf("HandleMessage() sentContent = %q, want %q", sender.sentContent, tt.wantContent)
				}
			}
		})
	}
}
