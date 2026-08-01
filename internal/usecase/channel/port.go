package channel

import "context"

// Creator はギルドに公開テキストチャンネルを作成するport。infrastructure層が実装する
type Creator interface {
	CreateTextChannel(ctx context.Context, guildID, name string) error
}

// ProposalMessenger はチャンネル作成提案メッセージの送信・自己リアクション付与を行うport。
// 戻り値のmessageIDは、後続のリアクション集計でどの提案に対する反応かを特定するために使う
type ProposalMessenger interface {
	SendProposal(ctx context.Context, channelID, content string) (messageID string, err error)
}

// ApprovalCounter はDiscord上の実際のリアクション状態から、提案メッセージに反応した
// ユニークユーザー数を数えるport。ユーザーがリアクションを取り消すケースを正しく扱うため、
// イベント側で受信した増分ではなく毎回Discordへ問い合わせて現在のユニーク人数を取得する
type ApprovalCounter interface {
	CountUniqueReactors(ctx context.Context, channelID, messageID string) (int, error)
}

// ProposalRepository はチャンネル作成提案の永続化を担うport。infrastructure層が実装する
type ProposalRepository interface {
	Save(ctx context.Context, p Proposal) error
	FindByMessage(ctx context.Context, channelID, messageID string) (Proposal, bool, error)
	MarkResolved(ctx context.Context, channelID, messageID string) error
}

// Proposal はチャンネル作成提案の永続化用データ
type Proposal struct {
	GuildID     string
	ChannelID   string
	MessageID   string
	ChannelName string
	ProposerID  string
	// Resolved は既にチャンネル作成が実行済み（可決処理が完了済み）かどうか。
	// リアクション増減のたびに複数回呼ばれ得るイベントハンドラから、同じ提案に対して
	// 二重にチャンネルを作成してしまうことを防ぐためのガードに使う
	Resolved bool
}

// SettingRepository はギルドごとの必要承認人数設定の永続化を担うport。infrastructure層が実装する
type SettingRepository interface {
	Get(ctx context.Context, guildID string) (requiredApprovals int, found bool, err error)
	Set(ctx context.Context, guildID string, requiredApprovals int) error
}
