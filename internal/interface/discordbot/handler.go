package discordbot

import (
	"context"
	"fmt"
	"strings"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/emoji"
)

// MessageSender はDiscordチャンネルへの返信を行うport
type MessageSender interface {
	SendMessage(ctx context.Context, channelID, content string) error
}

// IncomingMessage はdiscordgoのMessageCreateイベントを薄くラップした型。
// usecase層にdiscordgoの型が漏れないようにするための境界
type IncomingMessage struct {
	GuildID       string
	ChannelID     string
	AuthorID      string
	Content       string
	MentionsBotID bool   // Botへのメンションが含まれるか
	BotID         string // BotのユーザーID
	IsAdmin       bool   // 発言者がこのギルドで管理者権限を持つか
}

// mentionTag はbotIDから使い方案内文言に埋め込むDiscordの構造化メンション表記
// （"<@123456>"）を組み立てる。Bot名を固定文字列でハードコードすると環境（サーバー上の
// 表示名やBotアカウント自体）が変わった際に文言が実態とずれるため、実行時のBotIDから動的に組み立てる
func mentionTag(botID string) string {
	return fmt.Sprintf("<@%s>", botID)
}

// stripMentionTokens はスペース区切りのフィールド列から、構造化メンション
// （<@123456>等）を除去したフィールド列を返す。
// keyword/helpなど、メンションに続くコマンド本体を解析する各ハンドラで共通して使う
func stripMentionTokens(fields []string) []string {
	filtered := make([]string, 0, len(fields))
	for _, f := range fields {
		if strings.HasPrefix(f, "<@") && strings.HasSuffix(f, ">") {
			continue
		}
		filtered = append(filtered, f)
	}
	return filtered
}

// IncomingEmojiUpdate はdiscordgoのGuildEmojisUpdateイベントを薄くラップした型
type IncomingEmojiUpdate struct {
	GuildID     string
	AddedEmojis []emoji.Emoji
}

// MessageHandler はメッセージ受信時に処理を行うプラグインの契約。
// 新しいメッセージ系Bot機能はこのインターフェースを実装しrouterへ登録するだけで有効化される
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg IncomingMessage) error
}

// EmojiUpdateHandler はギルド絵文字更新時に処理を行うプラグインの契約
type EmojiUpdateHandler interface {
	HandleEmojiUpdate(ctx context.Context, ev IncomingEmojiUpdate) error
}
