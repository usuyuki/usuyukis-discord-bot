package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// reactionUsersLimit はリアクションユーザー一覧取得1回あたりの上限（Discord APIの最大値）
const reactionUsersLimit = 100

// ChannelProposalMessenger はusecase/channel.ProposalMessengerのdiscordgo実装
type ChannelProposalMessenger struct {
	session *discordgo.Session
}

// NewChannelProposalMessenger はChannelProposalMessengerを生成する
func NewChannelProposalMessenger(session *discordgo.Session) *ChannelProposalMessenger {
	return &ChannelProposalMessenger{session: session}
}

// proposalApprovalEmoji は提案メッセージにBotが最初に付ける絵文字リアクション。
// ユーザーはこの絵文字（または他の任意の絵文字）でリアクションすることで賛成を表明できる
const proposalApprovalEmoji = "✅"

// SendProposal はchannelIDへcontentを送信し、Bot自身が承認の意思表示としてリアクションを
// 1つ付けた上でメッセージIDを返す。提案者自身の賛成もこのBotのリアクションではなく、
// 実際に提案者が同じ絵文字でリアクションすることでカウントされる想定のため、
// Botのリアクション自体は承認者数のカウント対象から除外する（CountUniqueReactors側で対応）
func (m *ChannelProposalMessenger) SendProposal(ctx context.Context, channelID, content string) (string, error) {
	msg, err := m.session.ChannelMessageSend(channelID, content)
	if err != nil {
		return "", fmt.Errorf("discord: failed to send channel proposal: %w", err)
	}
	if err := m.session.MessageReactionAdd(channelID, msg.ID, proposalApprovalEmoji); err != nil {
		// ここで提案をDBに保存せずエラーを返すと、送信済みメッセージだけが残り
		// 誰にも反応されない孤立した提案になってしまう。呼び出し元へエラーを返す前に
		// メッセージ自体を削除し、送信の失敗として扱えるようにする
		if delErr := m.session.ChannelMessageDelete(channelID, msg.ID); delErr != nil {
			return "", fmt.Errorf("discord: failed to add initial reaction to proposal %s: %w (cleanup also failed: %v)", msg.ID, err, delErr)
		}
		return "", fmt.Errorf("discord: failed to add initial reaction to proposal %s: %w", msg.ID, err)
	}
	return msg.ID, nil
}

// ChannelApprovalCounter はusecase/channel.ApprovalCounterのdiscordgo実装
type ChannelApprovalCounter struct {
	session *discordgo.Session
}

// NewChannelApprovalCounter はChannelApprovalCounterを生成する
func NewChannelApprovalCounter(session *discordgo.Session) *ChannelApprovalCounter {
	return &ChannelApprovalCounter{session: session}
}

// CountUniqueReactors はchannelID/messageIDに付いた全ての絵文字リアクションを走査し、
// Bot自身を除いたユニークなユーザーID数を返す。同じユーザーが複数の絵文字でリアクションしても
// 1人としてカウントする。任意の絵文字を承認の意思表示として扱う仕様のため、
// 特定の絵文字1種類だけでなくメッセージに付いた全種類の絵文字を対象にする
func (c *ChannelApprovalCounter) CountUniqueReactors(ctx context.Context, channelID, messageID string) (int, error) {
	msg, err := c.session.ChannelMessage(channelID, messageID)
	if err != nil {
		return 0, fmt.Errorf("discord: failed to fetch proposal message %s: %w", messageID, err)
	}

	botID := ""
	if c.session.State.User != nil {
		botID = c.session.State.User.ID
	}

	uniqueUsers := map[string]bool{}
	for _, reaction := range msg.Reactions {
		// 1回のMessageReactions呼び出しでは最大reactionUsersLimit件しか取得できないため、
		// 1つの絵文字に100人を超えるリアクションが付いた場合でも全員を数えられるよう、
		// 最後に取得したユーザーIDをafterカーソルにして取得件数が上限未満になるまで
		// ページネーションする
		after := ""
		for {
			users, err := c.session.MessageReactions(channelID, messageID, reaction.Emoji.APIName(), reactionUsersLimit, "", after)
			if err != nil {
				return 0, fmt.Errorf("discord: failed to fetch reactors for proposal %s: %w", messageID, err)
			}
			for _, u := range users {
				if u.ID == botID {
					continue
				}
				uniqueUsers[u.ID] = true
			}
			if len(users) < reactionUsersLimit {
				break
			}
			after = users[len(users)-1].ID
		}
	}
	return len(uniqueUsers), nil
}
