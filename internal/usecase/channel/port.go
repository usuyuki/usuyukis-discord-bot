package channel

import "context"

// Creator はギルドへ新規テキストチャンネルを作成するport。infrastructure層が実装する。
// 作成と同時に、@everyoneには非表示・creatorUserIDのみ閲覧可能な権限オーバーライドを
// 設定することを実装側の責務とする（サーバー管理者はDiscordの仕様上チャンネル単位の
// オーバーライドを常にバイパスするため、明示的な許可設定は不要）
type Creator interface {
	CreatePrivateChannel(ctx context.Context, guildID, name, creatorUserID string) (channelID string, err error)
}
