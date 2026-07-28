package emoji

import (
	"context"
	"strings"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

// UseCase は絵文字追加の通知に関するアプリケーションロジック
type UseCase struct {
	channelFinder NotifyChannelFinder
	sender        MessageSender
}

// New はUseCaseを生成する
func New(channelFinder NotifyChannelFinder, sender MessageSender) *UseCase {
	return &UseCase{channelFinder: channelFinder, sender: sender}
}

// NotifyAdded は追加された絵文字名のリストを、ギルドに登録された通知先チャンネルへ通知する。
// 通知先が未登録の場合は何もしない（fallback先を持たない仕様）
func (u *UseCase) NotifyAdded(ctx context.Context, guildID string, addedEmojiNames []string) error {
	if len(addedEmojiNames) == 0 {
		return nil
	}

	nc, ok, err := u.channelFinder.Find(ctx, guildID, notifychannel.PurposeEmoji)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	content := "絵文字が追加されました: " + strings.Join(addedEmojiNames, ", ")
	return u.sender.SendMessage(ctx, nc.ChannelID, content)
}
