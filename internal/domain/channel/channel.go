// Package channel はDiscordのチャンネル権限オーバーライドに関する純粋な判定ロジックを提供する
package channel

// Discordの権限ビット値。domain層は外部ライブラリに依存しないため、discordgoの定数を
// 参照せずリテラルとして持つ（値そのものはDiscord API仕様として固定されている）
const (
	permViewChannel    int64 = 1 << 10
	permManageChannels int64 = 1 << 4
	permAdministrator  int64 = 1 << 3
)

// Overwrite はチャンネルに設定されたロール単位の権限オーバーライドを表す薄い値
type Overwrite struct {
	RoleID string
	Deny   int64
}

// IsPrivate はoverwritesの中に@everyoneロール（IDはguildIDと同一）へのViewChannel拒否が
// 含まれていればtrueを返す。Discordの「プライベートチャンネル」はこの状態を指す
func IsPrivate(guildID string, overwrites []Overwrite) bool {
	for _, ow := range overwrites {
		if ow.RoleID == guildID && ow.Deny&permViewChannel != 0 {
			return true
		}
	}
	return false
}

// IsChannelManagerRole はpermissionsがManageChannelsを持ちながらAdministratorは
// 持たないロールかどうかを判定する。Administrator権限保持者はDiscordの仕様上チャンネル単位の
// 権限オーバーライドを常にバイパスするため、明示的な閲覧拒否を設定しても操作できてしまい対象外となる。
// この関数がtrueを返すロールは、明示的にオーバーライドで拒否しない限り一般ユーザーでも
// 他人が作った非公開チャンネルへアクセスできてしまう対象を表す
func IsChannelManagerRole(permissions int64) bool {
	return permissions&permManageChannels != 0 && permissions&permAdministrator == 0
}
