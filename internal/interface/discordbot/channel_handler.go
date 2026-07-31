package discordbot

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	channelcreateUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/channelcreate"
)

// mentionTokenPattern はDiscordの構造化メンション表記（"<@id>"または（ニックネームメンション形式の）
// "<@!id>"）からユーザーIDを取り出す
var mentionTokenPattern = regexp.MustCompile(`^<@!?([^<>]+)>$`)

// ChannelHandler はチャンネル作成コマンド（@Bot channel create/create-private）を担うハンドラ。
// チャンネル作成自体はBot自身のManage Channels権限で行うため、コマンド実行者に管理者権限や
// チャンネル管理権限は不要（「ユーザーには管理者権限を渡したくないが、チャンネル作成は自由にさせたい」
// という運用要件のための機能）
type ChannelHandler struct {
	uc     *channelcreateUC.UseCase
	sender MessageSender
}

// NewChannelHandler はChannelHandlerを生成する
func NewChannelHandler(uc *channelcreateUC.UseCase, sender MessageSender) *ChannelHandler {
	return &ChannelHandler{uc: uc, sender: sender}
}

// HandleMessage はBotへの構造化メンションに続く本文が"channel create"/"channel create-private"の
// 場合にチャンネルを作成する
func (h *ChannelHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if !msg.MentionsBotID {
		return nil
	}
	cmd := parseChannelCommand(msg.Content, msg.BotID)
	if cmd == nil {
		return nil
	}
	if msg.GuildID == "" {
		return h.sender.SendMessage(ctx, msg.ChannelID, "このコマンドはサーバー内でのみ使用できます")
	}

	switch cmd.Sub {
	case "create":
		if cmd.Name == "" {
			return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("使い方: %s channel create <チャンネル名>", mentionTag(msg.BotID)))
		}
		return h.create(ctx, msg, cmd.Name, false, nil)

	case "create-private":
		if cmd.Name == "" {
			return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("使い方: %s channel create-private <チャンネル名> [@メンバー ...]", mentionTag(msg.BotID)))
		}
		return h.create(ctx, msg, cmd.Name, true, cmd.MemberIDs)

	default:
		return nil
	}
}

func (h *ChannelHandler) create(ctx context.Context, msg IncomingMessage, name string, private bool, memberIDs []string) error {
	channelID, err := h.uc.Create(ctx, msg.GuildID, name, private, msg.AuthorID, memberIDs)
	if err != nil {
		return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("チャンネル作成に失敗しました: %v", err))
	}
	visibility := "公開"
	if private {
		visibility = "プライベート"
	}
	return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("%sチャンネルを作成しました: <#%s>", visibility, channelID))
}

// channelCommand はパース済みのchannelコマンド引数
type channelCommand struct {
	Sub       string // create | create-private
	Name      string
	MemberIDs []string
}

// parseChannelCommand はメンションを除いたメッセージ本文から
// "channel create <name>" または "channel create-private <name> [@member ...]" 形式の
// コマンドを解析する。"channel"で始まらない場合はnilを返す（他のコマンドと区別するため）
func parseChannelCommand(content, botID string) *channelCommand {
	filtered := stripMentionTokens(strings.Fields(content), botID)
	if len(filtered) < 2 || filtered[0] != "channel" {
		return nil
	}

	cmd := &channelCommand{Sub: filtered[1]}
	switch cmd.Sub {
	case "create":
		if len(filtered) >= 3 {
			cmd.Name = filtered[2]
		}
	case "create-private":
		if len(filtered) >= 3 {
			cmd.Name = filtered[2]
		}
		if len(filtered) > 3 {
			cmd.MemberIDs = extractMentionedUserIDs(filtered[3:])
		}
	default:
		return nil
	}
	return cmd
}

// extractMentionedUserIDs はフィールド列から構造化メンション表記に一致するものだけを
// ユーザーIDへ変換して返す。一致しないフィールドは無視する
func extractMentionedUserIDs(fields []string) []string {
	var ids []string
	for _, f := range fields {
		m := mentionTokenPattern.FindStringSubmatch(f)
		if m == nil {
			continue
		}
		ids = append(ids, m[1])
	}
	return ids
}
