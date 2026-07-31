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
	if !msg.MentionsBotID || !isHelpCommand(msg.Content, msg.BotID) {
		return nil
	}
	return h.sender.SendMessage(ctx, msg.ChannelID, helpMessage(msg.BotID, msg.BotName))
}

// isHelpCommand はメンションを除いたメッセージ本文がちょうど"help"または"usage"（大文字小文字を無視）
// かどうかを判定する
func isHelpCommand(content, botID string) bool {
	filtered := stripMentionTokens(strings.Fields(content), botID)
	if len(filtered) != 1 {
		return false
	}
	word := strings.ToLower(filtered[0])
	return word == "help" || word == "usage"
}

// helpMessage はBotの機能一覧を説明する応答文言を組み立てる。
// README.mdの機能一覧表の内容に準拠しており、機能追加・変更時はREADME.mdと合わせて更新すること。
// help/usageコマンドの案内には実行時のBotIDから組み立てる構造化メンション表記を使う一方、
// キーワード等botIDに依存しないBot名表示にはbotNameをそのまま地の文で埋め込む。
// botNameが取得できなかった場合（State未同期など）はメンションタグにフォールバックする
func helpMessage(botID, botName string) string {
	tag := mentionTag(botID)
	name := botName
	if name == "" {
		name = tag
	}
	return fmt.Sprintf(
		"%sの機能一覧:\n"+
			"- キーワード自動応答: `%s keyword add <キーワード> <応答>` で登録したキーワードが発言に含まれると自動応答する（複数応答登録時はランダムで1つ選ばれる）。`keyword remove <キーワード> [応答]` で削除、`keyword list` で一覧表示。応答に `{$now}` を含めると現在日時に展開される（add/removeは管理者権限が必要）\n"+
			"- 俳句・短歌検知: 投稿を形態素解析し、5-7-5または5-7-5-7-7と判定できたら投稿元チャンネルへ通知する\n"+
			"- 絵文字追加通知: ギルドへ絵文字が追加されたことを検知して通知する\n"+
			"- スロット: 「%s」と発言する（メンション不要）とギルドのカスタム絵文字（少なければ標準絵文字）から3つ抽選する。3つ揃うと大当たり、2つ揃うと小当たり\n"+
			"- チャンネル作成: `%s channel create <チャンネル名>` で誰でもチャンネルを作成できる。作成したチャンネルは依頼者本人とサーバー管理者以外には見えない\n"+
			"- ヘルプ: `%s help` または `%s usage` でこの一覧を表示する",
		name, tag, slotTriggerPhrase, tag, tag, tag,
	)
}
