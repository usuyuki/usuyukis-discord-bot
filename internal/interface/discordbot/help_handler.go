package discordbot

import (
	"context"
	"fmt"
	"strings"
)

// HelpHandler はBotへのメンションで"help"/"usage"と言われた際に機能一覧を返すハンドラ
type HelpHandler struct {
	sender MessageSender
}

// NewHelpHandler はHelpHandlerを生成する
func NewHelpHandler(sender MessageSender) *HelpHandler {
	return &HelpHandler{sender: sender}
}

// HandleMessage はBotへのメンションに続く本文が"help"または"usage"の場合、機能一覧を返信する
func (h *HelpHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if !msg.MentionsBotID || !isHelpCommand(msg.Content) {
		return nil
	}
	return h.sender.SendMessage(ctx, msg.ChannelID, helpMessage(msg.BotID))
}

// isHelpCommand はメンションを除いたメッセージ本文がちょうど"help"または"usage"（大文字小文字を無視）
// かどうかを判定する
func isHelpCommand(content string) bool {
	filtered := stripMentionTokens(strings.Fields(content))
	if len(filtered) != 1 {
		return false
	}
	word := strings.ToLower(filtered[0])
	return word == "help" || word == "usage"
}

// helpMessage はBotの機能一覧を説明する応答文言を組み立てる。
// README.mdの機能一覧表の内容に準拠しており、機能追加・変更時はREADME.mdと合わせて更新すること。
// Botの表示名はサーバー・アカウントごとに変わり得るため固定文字列で埋め込まず、
// 実行時のBotIDから構造化メンション表記を組み立てて使う
func helpMessage(botID string) string {
	tag := mentionTag(botID)
	return fmt.Sprintf(
		"Botの機能一覧:\n"+
			"- キーワード自動応答: `%s keyword add <キーワード> <応答>` で登録したキーワードが発言に含まれると自動応答する（複数応答登録時はランダムで1つ選ばれる）。`keyword remove <キーワード> [応答]` で削除、`keyword list` で一覧表示。応答に `{$now}` を含めると現在日時に展開される（add/removeは管理者権限が必要）\n"+
			"- 俳句・短歌検知: 投稿を形態素解析し、5-7-5または5-7-5-7-7と判定できたら投稿元チャンネルへ通知する\n"+
			"- 絵文字追加通知: ギルドへ絵文字が追加されたことを検知して通知する\n"+
			"- ヘルプ: `%s help` または `%s usage` でこの一覧を表示する",
		tag, tag, tag,
	)
}
