package haiku

import (
	"context"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

type fakeAnalyzer struct {
	words []haiku.Word
	err   error
}

func (f *fakeAnalyzer) AnalyzeWords(text string) ([]haiku.Word, error) {
	return f.words, f.err
}

type fakeChannelFinder struct {
	nc notifychannel.NotifyChannel
	ok bool
}

func (f *fakeChannelFinder) Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error) {
	return f.nc, f.ok, nil
}

type fakeSender struct {
	sentChannelID string
	sentContent   string
	called        bool
}

func (f *fakeSender) SendMessage(ctx context.Context, channelID, content string) error {
	f.sentChannelID = channelID
	f.sentContent = content
	f.called = true
	return nil
}

func TestUseCase_Detect(t *testing.T) {
	haikuWords := []haiku.Word{
		{Surface: "古池", MoraCount: 3},
		{Surface: "や", MoraCount: 2},
		{Surface: "蛙", MoraCount: 3},
		{Surface: "飛び込む", MoraCount: 4},
		{Surface: "水の", MoraCount: 3},
		{Surface: "音", MoraCount: 2},
	}
	tankaWords := []haiku.Word{
		{Surface: "たら", MoraCount: 2},
		{Surface: "ちねの", MoraCount: 3},
		{Surface: "母が", MoraCount: 3},
		{Surface: "つりたる", MoraCount: 4},
		{Surface: "青蚊帳", MoraCount: 5},
		{Surface: "すがしい", MoraCount: 4},
		{Surface: "朝の", MoraCount: 3},
		{Surface: "かぜの", MoraCount: 3},
		{Surface: "中に", MoraCount: 2},
		{Surface: "ゐる", MoraCount: 2},
	}
	notWords := []haiku.Word{
		{Surface: "今日", MoraCount: 2},
		{Surface: "はいい", MoraCount: 3},
		{Surface: "天気", MoraCount: 3},
		{Surface: "ですね", MoraCount: 3},
	}

	tests := []struct {
		name              string
		words             []haiku.Word
		notifyChannelOK   bool
		notifyChannelID   string
		fallbackChannelID string
		wantDetected      bool
		wantSentChannelID string
		wantContent       string
	}{
		{
			name:              "正常系: 575判定でtrueかつ通知先チャンネル登録済みならそこに送信する",
			words:             haikuWords,
			notifyChannelOK:   true,
			notifyChannelID:   "notify-channel",
			fallbackChannelID: "fallback-channel",
			wantDetected:      true,
			wantSentChannelID: "notify-channel",
			wantContent:       "俳句を検知しました:\n古池や 蛙飛び込む 水の音",
		},
		{
			name:              "正常系: 通知先未登録ならfallbackチャンネルに送信する",
			words:             haikuWords,
			notifyChannelOK:   false,
			fallbackChannelID: "fallback-channel",
			wantDetected:      true,
			wantSentChannelID: "fallback-channel",
			wantContent:       "俳句を検知しました:\n古池や 蛙飛び込む 水の音",
		},
		{
			name:              "正常系: 57577判定でtrueなら短歌として通知する",
			words:             tankaWords,
			notifyChannelOK:   false,
			fallbackChannelID: "fallback-channel",
			wantDetected:      true,
			wantSentChannelID: "fallback-channel",
			wantContent:       "短歌を検知しました:\nたらちねの 母がつりたる 青蚊帳 すがしい朝の かぜの中にゐる",
		},
		{
			name:              "異常系: 575にも57577にもならなければ送信しない",
			words:             notWords,
			notifyChannelOK:   false,
			fallbackChannelID: "fallback-channel",
			wantDetected:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &fakeAnalyzer{words: tt.words}
			finder := &fakeChannelFinder{ok: tt.notifyChannelOK, nc: notifychannel.NotifyChannel{ChannelID: tt.notifyChannelID}}
			sender := &fakeSender{}
			u := New(analyzer, finder, sender)

			detected, err := u.Detect(context.Background(), "g1", tt.fallbackChannelID, "テストメッセージ")
			if err != nil {
				t.Fatalf("Detect() unexpected error = %v", err)
			}
			if detected != tt.wantDetected {
				t.Fatalf("Detect() = %v, want %v", detected, tt.wantDetected)
			}
			if tt.wantDetected {
				if !sender.called {
					t.Fatal("Detect() expected SendMessage to be called, but it wasn't")
				}
				if sender.sentChannelID != tt.wantSentChannelID {
					t.Errorf("Detect() sentChannelID = %q, want %q", sender.sentChannelID, tt.wantSentChannelID)
				}
				if sender.sentContent != tt.wantContent {
					t.Errorf("Detect() sentContent = %q, want %q", sender.sentContent, tt.wantContent)
				}
			} else if sender.called {
				t.Error("Detect() expected SendMessage not to be called, but it was")
			}
		})
	}
}
