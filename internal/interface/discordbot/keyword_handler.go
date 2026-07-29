package discordbot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
	keywordUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/keyword"
)

// KeywordHandler はキーワード登録コマンド（@Bot keyword add/remove/list）と
// 通常メッセージへの自動応答を担うハンドラ。
// 応答文言に含まれる{$now}プレースホルダーはnowが返す現在時刻に展開される（打刻Bot廃止後、
// 時刻を含む応答もこのキーワード自動応答経由で登録する運用に統一している）
type KeywordHandler struct {
	uc     *keywordUC.UseCase
	sender MessageSender
	now    func() time.Time
}

// NewKeywordHandler はKeywordHandlerを生成する
func NewKeywordHandler(uc *keywordUC.UseCase, sender MessageSender) *KeywordHandler {
	return &KeywordHandler{uc: uc, sender: sender, now: time.Now}
}

// HandleMessage はBotへの構造化メンション（MentionsBotID）であればコマンドとして解釈し、
// そうでなければ通常メッセージとして登録済みキーワードとのマッチを試みる
func (h *KeywordHandler) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	if msg.MentionsBotID {
		return h.handleCommand(ctx, msg)
	}
	return h.handleAutoReply(ctx, msg)
}

func (h *KeywordHandler) handleCommand(ctx context.Context, msg IncomingMessage) error {
	args := parseKeywordCommand(msg.Content)
	if args == nil {
		return nil
	}

	switch args.Sub {
	case "add":
		if !msg.IsAdmin {
			return h.sender.SendMessage(ctx, msg.ChannelID, "キーワード登録には管理者権限が必要です")
		}
		if args.Word == "" || args.Response == "" {
			return h.sender.SendMessage(ctx, msg.ChannelID, "使い方: @Bot keyword add <キーワード> <応答>")
		}
		if err := h.uc.Register(ctx, msg.GuildID, args.Word, args.Response); err != nil {
			return err
		}
		return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("登録しました: %s → %s", args.Word, args.Response))

	case "remove":
		if !msg.IsAdmin {
			return h.sender.SendMessage(ctx, msg.ChannelID, "キーワード削除には管理者権限が必要です")
		}
		if args.Word == "" {
			return h.sender.SendMessage(ctx, msg.ChannelID, "使い方: @Bot keyword remove <キーワード> [応答]")
		}
		// 応答を指定した場合はその応答のみ削除し、未指定の場合はキーワードごと全応答を削除する
		if args.Response == "" {
			if err := h.uc.RemoveKeyword(ctx, msg.GuildID, args.Word); err != nil {
				return err
			}
			return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("削除しました: %s", args.Word))
		}
		if err := h.uc.RemoveResponse(ctx, msg.GuildID, args.Word, args.Response); err != nil {
			return err
		}
		return h.sender.SendMessage(ctx, msg.ChannelID, fmt.Sprintf("削除しました: %s → %s", args.Word, args.Response))

	case "list":
		keywords, err := h.uc.List(ctx, msg.GuildID)
		if err != nil {
			return err
		}
		return h.sender.SendMessage(ctx, msg.ChannelID, formatKeywordList(keywords))

	default:
		return nil
	}
}

func (h *KeywordHandler) handleAutoReply(ctx context.Context, msg IncomingMessage) error {
	k, ok, err := h.uc.Match(ctx, msg.GuildID, msg.Content)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return h.sender.SendMessage(ctx, msg.ChannelID, k.RandomResponse(h.now()))
}

// keywordCommand はパース済みのkeywordコマンド引数
type keywordCommand struct {
	Sub      string // add | remove | list
	Word     string
	Response string
}

// parseKeywordCommand はメンションを除いたメッセージ本文から
// "keyword add|remove|list ..." 形式のコマンドを解析する。
// "keyword"で始まらない場合はnilを返す（他のコマンドやただのメンションと区別するため）
func parseKeywordCommand(content string) *keywordCommand {
	fields := strings.Fields(content)
	// メンション文字列（<@123456>等）を除去してからkeywordコマンドを探す
	filtered := make([]string, 0, len(fields))
	for _, f := range fields {
		if strings.HasPrefix(f, "<@") && strings.HasSuffix(f, ">") {
			continue
		}
		filtered = append(filtered, f)
	}
	if len(filtered) == 0 || filtered[0] != "keyword" {
		return nil
	}
	if len(filtered) < 2 {
		return nil
	}

	cmd := &keywordCommand{Sub: filtered[1]}
	switch cmd.Sub {
	case "add":
		if len(filtered) >= 3 {
			cmd.Word = filtered[2]
		}
		if len(filtered) >= 4 {
			cmd.Response = strings.Join(filtered[3:], " ")
		}
	case "remove":
		if len(filtered) >= 3 {
			cmd.Word = filtered[2]
		}
		if len(filtered) >= 4 {
			cmd.Response = strings.Join(filtered[3:], " ")
		}
	case "list":
		// 追加引数なし
	default:
		return nil
	}
	return cmd
}

func formatKeywordList(keywords []keyword.Keyword) string {
	if len(keywords) == 0 {
		return "登録済みキーワードはありません"
	}
	var b strings.Builder
	b.WriteString("登録済みキーワード:\n")
	for _, k := range keywords {
		fmt.Fprintf(&b, "- %s → %s\n", k.Word, strings.Join(k.Responses, " / "))
	}
	return strings.TrimRight(b.String(), "\n")
}
