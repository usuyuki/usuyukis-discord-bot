package discordbot

import (
	"context"
	"strings"
	"testing"
)

func TestHelpHandler_HandleMessage(t *testing.T) {
	tests := []struct {
		name       string
		msg        IncomingMessage
		wantCalled bool
	}{
		{
			name:       "正常系: 構造化メンションに続けてhelpと言うと機能一覧を返信する",
			msg:        IncomingMessage{ChannelID: "c1", BotID: "bot", BotName: "usuyuki", MentionsBotID: true, Content: "<@bot> help"},
			wantCalled: true,
		},
		{
			name:       "正常系: 構造化メンションに続けてusageと言うと機能一覧を返信する",
			msg:        IncomingMessage{ChannelID: "c1", BotID: "bot", BotName: "usuyuki", MentionsBotID: true, Content: "<@bot> usage"},
			wantCalled: true,
		},
		{
			name:       "異常系: 構造化メンションでなくテキストの@メンション風表記のみでは反応しない",
			msg:        IncomingMessage{ChannelID: "c1", MentionsBotID: false, Content: "@bot help"},
			wantCalled: false,
		},
		{
			name:       "正常系: 大文字小文字を無視してHelp/USAGEも認識する",
			msg:        IncomingMessage{ChannelID: "c1", BotID: "bot", MentionsBotID: true, Content: "<@bot> HELP"},
			wantCalled: true,
		},
		{
			name:       "異常系: メンションがないと反応しない",
			msg:        IncomingMessage{ChannelID: "c1", MentionsBotID: false, Content: "help"},
			wantCalled: false,
		},
		{
			name:       "異常系: メンションがあってもhelp/usage以外の本文には反応しない",
			msg:        IncomingMessage{ChannelID: "c1", MentionsBotID: true, Content: "<@bot> keyword list"},
			wantCalled: false,
		},
		{
			name:       "異常系: helpに余分な引数が付くと反応しない",
			msg:        IncomingMessage{ChannelID: "c1", MentionsBotID: true, Content: "<@bot> help me"},
			wantCalled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeMessageSender{}
			h := NewHelpHandler(sender)

			if err := h.HandleMessage(context.Background(), tt.msg); err != nil {
				t.Fatalf("HandleMessage() unexpected error = %v", err)
			}
			if sender.called != tt.wantCalled {
				t.Fatalf("HandleMessage() called = %v, want %v", sender.called, tt.wantCalled)
			}
			if tt.wantCalled {
				if sender.sentChannelID != tt.msg.ChannelID {
					t.Errorf("HandleMessage() sentChannelID = %q, want %q", sender.sentChannelID, tt.msg.ChannelID)
				}
				if !strings.Contains(sender.sentContent, "キーワード自動応答") {
					t.Errorf("HandleMessage() content = %q, want to contain feature list", sender.sentContent)
				}
				if !strings.Contains(sender.sentContent, "チャンネル作成") {
					t.Errorf("HandleMessage() content = %q, want to contain channel creation feature", sender.sentContent)
				}
				wantTag := mentionTag(tt.msg.BotID)
				if !strings.Contains(sender.sentContent, wantTag) {
					t.Errorf("HandleMessage() content = %q, want to contain dynamically built mention tag %q", sender.sentContent, wantTag)
				}
				if !strings.Contains(sender.sentContent, tt.msg.BotName) {
					t.Errorf("HandleMessage() content = %q, want to contain bot name %q", sender.sentContent, tt.msg.BotName)
				}
			}
		})
	}
}

func TestIsHelpCommand(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "正常系: メンション除去後にhelpのみが残れば真",
			content: "<@bot> help",
			want:    true,
		},
		{
			name:    "正常系: メンション除去後にusageのみが残れば真",
			content: "<@bot> usage",
			want:    true,
		},
		{
			name:    "異常系: 余分な引数が付くと偽",
			content: "<@bot> help me",
			want:    false,
		},
		{
			name:    "異常系: help/usage以外の単語は偽",
			content: "<@bot> keyword",
			want:    false,
		},
		{
			name:    "異常系: 本文が空の場合は偽",
			content: "<@bot>",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHelpCommand(tt.content, "bot"); got != tt.want {
				t.Errorf("isHelpCommand(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}
