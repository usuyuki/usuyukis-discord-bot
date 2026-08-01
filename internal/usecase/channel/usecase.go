package channel

import (
	"context"
	"fmt"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channel"
)

// UseCase は一般ユーザーからのコマンドに応じてBotが代理でチャンネルを作成するアプリケーションロジック。
// 一般ユーザーにManageChannelsロールを付与する運用は、ギルド全体に及ぶ権限のため他人の
// 非公開チャンネルまで操作できてしまう問題があった（adr/0012, adr/0013参照）。この問題を
// 構造的に避けるため、チャンネル作成はBot自身の権限で代行する。
// さらにチャンネルの乱立を防ぐため即時作成はせず、提案メッセージへのリアクションが
// ギルドごとの必要承認人数（デフォルト2、提案者自身を含む）に達した時点で作成する
// 投票制フローとする（adr/0014参照）
type UseCase struct {
	creator   Creator
	messenger ProposalMessenger
	counter   ApprovalCounter
	proposals ProposalRepository
	settings  SettingRepository
}

// New はUseCaseを生成する
func New(creator Creator, messenger ProposalMessenger, counter ApprovalCounter, proposals ProposalRepository, settings SettingRepository) *UseCase {
	return &UseCase{creator: creator, messenger: messenger, counter: counter, proposals: proposals, settings: settings}
}

// Propose はnameを検証した上で、channelIDに提案メッセージを送信し、提案を保存する。
// この時点ではチャンネルはまだ作成しない
func (u *UseCase) Propose(ctx context.Context, guildID, channelID, proposerID, name string) error {
	validName, err := channel.NewName(name)
	if err != nil {
		return err
	}

	required, err := u.requiredApprovals(ctx, guildID)
	if err != nil {
		return err
	}
	content := fmt.Sprintf("#%s を作ろうとしています…\n✅で%d人以上のリアクションが集まると作成が可決されます！", validName.String(), required.Int())
	messageID, err := u.messenger.SendProposal(ctx, channelID, content)
	if err != nil {
		return err
	}

	return u.proposals.Save(ctx, Proposal{
		GuildID:     guildID,
		ChannelID:   channelID,
		MessageID:   messageID,
		ChannelName: validName.String(),
		ProposerID:  proposerID,
	})
}

// RecordReaction は提案メッセージへのリアクション追加をきっかけに、現在の承認者数を
// 数え直し、必要承認人数に達していればチャンネルを作成し提案を解決済みにする。
// 対象の提案が見つからない、または既に解決済みの場合は何もしない
func (u *UseCase) RecordReaction(ctx context.Context, channelID, messageID string) error {
	proposal, found, err := u.proposals.FindByMessage(ctx, channelID, messageID)
	if err != nil {
		return err
	}
	if !found || proposal.Resolved {
		return nil
	}

	required, err := u.requiredApprovals(ctx, proposal.GuildID)
	if err != nil {
		return err
	}
	count, err := u.counter.CountUniqueReactors(ctx, channelID, messageID)
	if err != nil {
		return err
	}
	if !channel.IsApproved(count, required.Int()) {
		return nil
	}

	if err := u.creator.CreateTextChannel(ctx, proposal.GuildID, proposal.ChannelName); err != nil {
		return err
	}
	return u.proposals.MarkResolved(ctx, channelID, messageID)
}

// requiredApprovals はguildIDに設定された必要承認人数を返す。未設定の場合はデフォルト値を使う
func (u *UseCase) requiredApprovals(ctx context.Context, guildID string) (channel.RequiredApprovals, error) {
	raw, found, err := u.settings.Get(ctx, guildID)
	if err != nil {
		return channel.RequiredApprovals{}, err
	}
	if !found {
		return channel.DefaultRequiredApprovals(), nil
	}
	return channel.NewRequiredApprovals(raw)
}

// GetRequiredApprovals は管理画面向けに、guildIDに設定された必要承認人数を返す。
// 未設定の場合はデフォルト値を返す
func (u *UseCase) GetRequiredApprovals(ctx context.Context, guildID string) (int, error) {
	required, err := u.requiredApprovals(ctx, guildID)
	if err != nil {
		return 0, err
	}
	return required.Int(), nil
}

// SetRequiredApprovals は管理画面からの入力で、guildIDの必要承認人数を設定する
func (u *UseCase) SetRequiredApprovals(ctx context.Context, guildID string, requiredApprovals int) error {
	validated, err := channel.NewRequiredApprovals(requiredApprovals)
	if err != nil {
		return err
	}
	return u.settings.Set(ctx, guildID, validated.Int())
}
