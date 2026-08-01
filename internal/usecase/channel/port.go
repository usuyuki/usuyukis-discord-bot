package channel

import "context"

// Restrictor はプライベートチャンネルに対し、チャンネル管理ロール（一般ユーザーに付与された
// ManageChannelsを持つロールでAdministratorは持たないもの）のアクセスを剥奪し、
// creatorUserIDのみ明示的に許可するport。infrastructure層が実装する。
// サーバー管理者（Administrator権限保持者）はDiscordの仕様上チャンネル単位の
// オーバーライドを常にバイパスするため、実装側で管理者向けの許可を追加する必要はない
type Restrictor interface {
	RestrictToCreatorAndAdmins(ctx context.Context, guildID, channelID, creatorUserID string) error
}
