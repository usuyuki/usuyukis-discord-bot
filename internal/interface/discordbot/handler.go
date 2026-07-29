package discordbot

import (
	"context"
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

// MentionsRakuro はdiscordgoの構造化メンション（MentionsBotID）と、
// 本文中のテキストとしての"@Rakuro"/"@rakuro"表記の両方を統一的に判定する。
// 各ハンドラ（打刻・keyword・俳句）はこの結果を使ってBot起動条件・二重応答回避を判定するため、
// ハンドラごとに判定基準がずれて二重応答や無反応が起きないようにここへ集約している
func (m IncomingMessage) MentionsRakuro() bool {
	return m.MentionsBotID || mentionsRakuroText(m.Content)
}

// mentionsRakuroText は本文中に単語として"@Rakuro"/"@rakuro"が含まれるかを
// 大文字小文字を無視して判定する。strings.Containsによる単純部分一致だと
// URLや無関係な単語（"@Rakuroski"等）内の文字列にも誤反応するため、
// 前後が英数字・アンダースコアで連結されていない（単語境界である）ことを確認する
func mentionsRakuroText(content string) bool {
	lower := strings.ToLower(content)
	const target = "@rakuro"
	idx := 0
	for {
		pos := strings.Index(lower[idx:], target)
		if pos == -1 {
			return false
		}
		start := idx + pos
		end := start + len(target)
		if isWordBoundaryRune(lower, start-1) && isWordBoundaryRune(lower, end) {
			return true
		}
		idx = start + 1
	}
}

// isWordBoundaryRune はlowerのbyte位置posにある文字（範囲外なら境界とみなす）が
// 英数字・アンダースコア以外かどうかを判定する
func isWordBoundaryRune(lower string, pos int) bool {
	if pos < 0 || pos >= len(lower) {
		return true
	}
	c := lower[pos]
	isWordChar := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
	return !isWordChar
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
