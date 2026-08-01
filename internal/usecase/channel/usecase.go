package channel

import "context"

// UseCase はチャンネル作成時のプライベートチャンネル保護に関するアプリケーションロジック。
// 一般ユーザーにも自らチャンネルを作れるようManageChannelsを持つロールを付与する運用を前提とするが、
// この権限はギルド全体の全チャンネルに及ぶため、そのままでは他人が作った既存の非公開チャンネルまで
// 閲覧・管理できてしまう。新規チャンネル作成イベントを検知しプライベートであれば都度そのチャンネルに
// 限定してロールのアクセスを剥奪することで、「チャンネル作成はできるが他人の非公開チャンネルは
// 操作できない」を実現する
type UseCase struct {
	restrictor Restrictor
}

// New はUseCaseを生成する
func New(restrictor Restrictor) *UseCase {
	return &UseCase{restrictor: restrictor}
}

// ProtectIfPrivate はisPrivateがtrueの場合のみ、チャンネル管理ロールのアクセスを剥奪し
// creatorUserIDのみ操作できる状態にする。プライベートでないチャンネルには何もしない
func (u *UseCase) ProtectIfPrivate(ctx context.Context, guildID, channelID, creatorUserID string, isPrivate bool) error {
	if !isPrivate {
		return nil
	}
	return u.restrictor.RestrictToCreatorAndAdmins(ctx, guildID, channelID, creatorUserID)
}
