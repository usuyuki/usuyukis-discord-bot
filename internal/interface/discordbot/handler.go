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
	BotName       string // Botの表示名（ヘルプ文言などの案内表示に使う）
	IsAdmin       bool   // 発言者がこのギルドで管理者権限を持つか
}

// mentionTag はbotIDから使い方案内文言に埋め込むDiscordの構造化メンション表記
// （"<@123456>"）を組み立てる。Bot名を固定文字列でハードコードすると環境（サーバー上の
// 表示名やBotアカウント自体）が変わった際に文言が実態とずれるため、実行時のBotIDから動的に組み立てる
func mentionTag(botID string) string {
	return fmt.Sprintf("<@%s>", botID)
}

// stripMentionTokens はスペース区切りのフィールド列から、指定したbotIDの構造化メンション
// （"<@botID>"）にちょうど一致するフィールドを全て除去し、さらに先頭フィールドがbotName
// （大文字小文字を無視）にちょうど一致する場合はそれも除去したフィールド列を返す。
// "<@...>"の形を持つ全フィールドを対象にすると、キーワードの値として渡された
// リテラルな"<@notanid>"のような引数まで誤って除去してしまうため、実際のBotメンションのみに絞る。
// 後者は、過去メッセージのコピペ等で構造化メンションではなく地の文の"@botName"表記に
// なってしまった場合でもコマンドとして認識できるようにするための救済措置（先頭のみ対象にし、
// 文中に偶然botNameと同じ単語が現れても誤除去しないようにする）。
// keyword/helpなど、メンションに続くコマンド本体を解析する各ハンドラで共通して使う
func stripMentionTokens(fields []string, botID, botName string) []string {
	tag := mentionTag(botID)
	filtered := make([]string, 0, len(fields))
	for _, f := range fields {
		if f == tag {
			continue
		}
		filtered = append(filtered, f)
	}
	if botName != "" && len(filtered) > 0 {
		first := strings.TrimPrefix(filtered[0], "@")
		if strings.EqualFold(first, botName) {
			filtered = filtered[1:]
		}
	}
	return filtered
}

// IncomingEmojiUpdate はdiscordgoのGuildEmojisUpdateイベントを薄くラップした型
type IncomingEmojiUpdate struct {
	GuildID     string
	AddedEmojis []emoji.Emoji
}

// IncomingReactionAdd はdiscordgoのMessageReactionAddイベントを薄くラップした型
type IncomingReactionAdd struct {
	ChannelID string
	MessageID string
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

// ReactionAddHandler はメッセージへのリアクション追加時に処理を行うプラグインの契約
type ReactionAddHandler interface {
	HandleReactionAdd(ctx context.Context, ev IncomingReactionAdd) error
}
