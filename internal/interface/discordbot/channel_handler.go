package discordbot

import (
	"context"
	"fmt"
	"strings"
)

// ChannelProposeUseCase はチャンネル作成提案コマンドが呼び出すport
type ChannelProposeUseCase interface {
	Propose(ctx context.Context, guildID, channelID, proposerID, name string) error
}

// ChannelHandler はチャンネル作成提案コマンド（@Bot channel create <name>）を担うハンドラ。
// 一般ユーザーにManageChannelsロールを付与する運用は他人の非公開チャンネルまで操作できて
// しまう問題があった（adr/0013参照）ため、チャンネル作成はBot自身の権限で代行する。
// 乱立防止のため即時作成はせず、UseCase.Proposeが送信する提案メッセージへの
// リアクションが必要人数集まった時点で作成される（ReactionHandlerが担当、adr/0014参照）
type ChannelHandler struct {
	uc     ChannelProposeUseCase
	sender MessageSender
}

// NewChannelHandler はChannelHandlerを生成する
func NewChannelHandler(uc ChannelProposeUseCase, sender MessageSender) *ChannelHandler {
	return &ChannelHandler{uc: uc, sender: sender}
}

// HandleMessage はBotへの構造化メンションに続く本文が"channel create <name>"の場合、
// チャンネル作成の提案を行う。提案が受理された場合はUseCase側が提案メッセージを
// 送信するため、ここから重ねて成功メッセージは送らない
func (h *ChannelHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if !msg.MentionsBotID {
		return nil
	}
	name, ok := parseChannelCreateCommand(msg.Content, msg.BotID, msg.BotMentionNames)
	if !ok {
		return nil
	}
	if name == "" {
		return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("使い方: %s channel create <チャンネル名>", mentionTag(msg.BotID)))
	}

	if err := h.uc.Propose(ctx, msg.GuildID, msg.ChannelID, msg.AuthorID, name); err != nil {
		// チャンネル名のバリデーションエラーなどユーザーの入力ミスに起因することが
		// 多いため、ハンドラのエラーとして扱わずチャンネルへ通知するだけにとどめる
		return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("チャンネルを提案できませんでした: %v", err))
	}
	return nil
}

// parseChannelCreateCommand はメンションを除いたメッセージ本文から
// "channel create <name>"形式のコマンドを解析する。"channel create"で始まらない場合は
// okがfalseになる。nameが省略された場合はokがtrueのまま空文字を返す（使い方案内の対象）
func parseChannelCreateCommand(content, botID string, botNames []string) (name string, ok bool) {
	filtered := stripMentionTokens(strings.Fields(content), botID, botNames)
	if len(filtered) < 2 || filtered[0] != "channel" || filtered[1] != "create" {
		return "", false
	}
	if len(filtered) < 3 {
		return "", true
	}
	return filtered[2], true
}
