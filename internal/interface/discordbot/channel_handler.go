package discordbot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channel"
	channelUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/channel"
)

// ChannelHandler は"@Bot channel create <チャンネル名>"コマンドを受け、ギルドに新規チャンネルを
// 作成するハンドラ。Discordの生の「チャンネルを管理」権限を一般ユーザーへ付与すると、新規作成だけでなく
// 既存の非公開チャンネルの閲覧・削除・移動まで許してしまうため、代わりにBot自身がこのコマンド経由で
// チャンネル作成のみを代行する。管理者権限チェックは行わず、全ユーザーが利用できる
type ChannelHandler struct {
	uc     *channelUC.UseCase
	sender MessageSender
}

// NewChannelHandler はChannelHandlerを生成する
func NewChannelHandler(uc *channelUC.UseCase, sender MessageSender) *ChannelHandler {
	return &ChannelHandler{uc: uc, sender: sender}
}

// HandleMessage はBotへの構造化メンションに続く本文が"channel create <チャンネル名>"であれば
// チャンネルを作成する。作成されたチャンネルは依頼者本人とサーバー管理者以外には見えない
func (h *ChannelHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if !msg.MentionsBotID {
		return nil
	}
	args := parseChannelCommand(msg.Content, msg.BotID)
	if args == nil {
		return nil
	}

	switch args.Sub {
	case "create":
		if args.Name == "" {
			return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("使い方: %s channel create <チャンネル名>", mentionTag(msg.BotID)))
		}
		channelID, err := h.uc.Create(ctx, msg.GuildID, args.Name, msg.AuthorID)
		if err != nil {
			if errors.Is(err, channel.ErrEmptyName) || errors.Is(err, channel.ErrNameTooLong) {
				return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("チャンネル名が不正です: %v", err))
			}
			return err
		}
		return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("チャンネルを作成しました: <#%s>（あなたとサーバー管理者以外には見えません）", channelID))

	default:
		return nil
	}
}

// channelCommand はパース済みのchannelコマンド引数
type channelCommand struct {
	Sub  string // create
	Name string
}

// parseChannelCommand はメンションを除いたメッセージ本文から"channel create <チャンネル名>"形式の
// コマンドを解析する。"channel"で始まらない場合はnilを返す
func parseChannelCommand(content, botID string) *channelCommand {
	filtered := stripMentionTokens(strings.Fields(content), botID)
	if len(filtered) == 0 || filtered[0] != "channel" {
		return nil
	}
	if len(filtered) < 2 {
		return nil
	}

	cmd := &channelCommand{Sub: filtered[1]}
	switch cmd.Sub {
	case "create":
		if len(filtered) >= 3 {
			cmd.Name = filtered[2]
		}
	default:
		return nil
	}
	return cmd
}
