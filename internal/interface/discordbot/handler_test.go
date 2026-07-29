package discordbot

import (
	"context"
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
