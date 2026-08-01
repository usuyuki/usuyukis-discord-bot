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
	GuildID         string
	ChannelID       string
	AuthorID        string
	Content         string
	MentionsBotID   bool     // Botへのメンションが含まれるか
	BotID           string   // BotのユーザーID
	BotName         string   // Botの表示名（ヘルプ文言などの案内表示に使う。Usernameを優先）
	BotMentionNames []string // 地の文@BotName救済判定で許容するBot名候補（Username/GlobalNameなど）
	IsAdmin         bool     // 発言者がこのギルドで管理者権限を持つか
}

// mentionTag はbotIDから使い方案内文言に埋め込むDiscordの構造化メンション表記
// （"<@123456>"）を組み立てる。Bot名を固定文字列でハードコードすると環境（サーバー上の
// 表示名やBotアカウント自体）が変わった際に文言が実態とずれるため、実行時のBotIDから動的に組み立てる
func mentionTag(botID string) string {
	return fmt.Sprintf("<@%s>", botID)
}

// mentionAtPrefixes はDiscordのメンション表記として使われる"@"の許容文字種。
// 半角"@"（U+0040）に加え、日本語IME/モバイル入力で混入しやすい全角"＠"（U+FF20）も
// 同一視することで、コピペ救済ロジックが入力方式に依存して失敗しないようにする
var mentionAtPrefixes = []string{"@", "＠"}

// trimMentionAtPrefix は先頭の半角/全角"@"を除去する。地の文@BotName救済判定を行う
// 箇所（detectsMentionsBotとstripMentionTokens）で共通して使うことで、両者の判定基準が
// 将来のいずれか片方だけの修正でズレることを防ぐ
func trimMentionAtPrefix(field string) string {
	for _, at := range mentionAtPrefixes {
		if trimmed := strings.TrimPrefix(field, at); trimmed != field {
			return trimmed
		}
	}
	return field
}

// MatchesBotMentionName はfieldが先頭"@"（全角含む）を除いた上でbotNamesのいずれか
// （大文字小文字を無視）に一致するかを判定する。infrastructure/discord側のdetectsMentionsBot
// からも参照されるためexportし、判定基準の重複実装によるズレを防ぐ
func MatchesBotMentionName(field string, botNames []string) bool {
	trimmed := trimMentionAtPrefix(field)
	for _, name := range botNames {
		if name != "" && strings.EqualFold(trimmed, name) {
			return true
		}
	}
	return false
}

// stripMentionTokens はスペース区切りのフィールド列から、指定したbotIDの構造化メンション
// （"<@botID>"）にちょうど一致するフィールドを全て除去したフィールド列を返す。
// 構造化メンションが1つも見つからなかった場合に限り、先頭フィールドがbotNamesのいずれか
// （大文字小文字を無視）にちょうど一致すればそれも除去する。
// "<@...>"の形を持つ全フィールドを対象にすると、キーワードの値として渡された
// リテラルな"<@notanid>"のような引数まで誤って除去してしまうため、実際のBotメンションのみに絞る。
// 後者は、過去メッセージのコピペ等で構造化メンションではなく地の文の"@botName"表記に
// なってしまった場合でもコマンドとして認識できるようにするための救済措置（先頭のみ対象にし、
// 文中に偶然botNameと同じ単語が現れても誤除去しないようにする）。構造化メンションを除去済みの
// 場合にまで平文フォールバックを重ねて適用すると、"<@botID> @someone ..."のように構造化メンション
// の直後に平文の"@BotName"風トークンが続く入力で、ユーザーが意図した引数まで誤って消えてしまうため、
// 両方を同時に適用しないようにする。
// keyword/helpなど、メンションに続くコマンド本体を解析する各ハンドラで共通して使う
func stripMentionTokens(fields []string, botID string, botNames []string) []string {
	tag := mentionTag(botID)
	filtered := make([]string, 0, len(fields))
	strippedStructuredMention := false
	for _, f := range fields {
		if f == tag {
			strippedStructuredMention = true
			continue
		}
		filtered = append(filtered, f)
	}
	if !strippedStructuredMention && len(filtered) > 0 && MatchesBotMentionName(filtered[0], botNames) {
		filtered = filtered[1:]
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
