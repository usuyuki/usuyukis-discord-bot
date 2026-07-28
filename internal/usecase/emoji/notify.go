package emoji

import (
	"context"
	"strings"

	domainEmoji "github.com/usuyuki/usuyukis-discord-bot/internal/domain/emoji"
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

// NotifyAdded は追加された絵文字のリストを、ギルドに登録された通知先チャンネルへ通知する。
// 通知先が未登録の場合は何もしない（fallback先を持たない仕様）
func (u *UseCase) NotifyAdded(ctx context.Context, guildID string, addedEmojis []domainEmoji.Emoji) error {
	if len(addedEmojis) == 0 {
		return nil
	}

	nc, ok, err := u.channelFinder.Find(ctx, guildID, notifychannel.PurposeEmoji)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	// 絵文字画像自体が見えるよう、名前だけでなくタグ（<:name:id>形式）も併記する
	descriptions := make([]string, 0, len(addedEmojis))
	for _, e := range addedEmojis {
		descriptions = append(descriptions, e.Tag()+" "+e.Name())
	}
	content := "新しい絵文字が追加されたぱか: " + strings.Join(descriptions, ", ")
	return u.sender.SendMessage(ctx, nc.ChannelID, content)
}
