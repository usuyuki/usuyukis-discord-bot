package channelcreate

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// maxNameLength はDiscordのチャンネル名文字数上限
const maxNameLength = 100

var (
	// ErrEmptyName は空文字（または空白のみ）のチャンネル名が渡された場合に返す
	ErrEmptyName = errors.New("channelcreate: name must not be empty")
	// ErrNameTooLong はDiscordのチャンネル名文字数上限を超えた場合に返す
	ErrNameTooLong = errors.New("channelcreate: name is too long")
	// ErrEmptyCreatorID は作成者のユーザーIDが空文字の場合に返す
	ErrEmptyCreatorID = errors.New("channelcreate: creatorID must not be empty")
)

// ChannelCreation はチャンネル作成コマンドの入力を検証済みの状態で表す値オブジェクト。
// Privateがtrueの場合、CreatorIDおよびMemberIDsに含まれるユーザーのみが閲覧できるチャンネルとして
// 作成する（Discordのチャンネル管理権限を持たないユーザーでも、Bot自身の権限でプライベートチャンネルを
// 作れるようにするための機能）
type ChannelCreation struct {
	Name      string
	Private   bool
	CreatorID string
	MemberIDs []string // Private時に閲覧を許可する、作成者以外の追加ユーザーID一覧（重複・作成者自身は除去済み）
}

// New は入力を検証してChannelCreationを生成する
func New(name string, private bool, creatorID string, memberIDs []string) (ChannelCreation, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ChannelCreation{}, ErrEmptyName
	}
	if utf8.RuneCountInString(trimmed) > maxNameLength {
		return ChannelCreation{}, ErrNameTooLong
	}
	if creatorID == "" {
		return ChannelCreation{}, ErrEmptyCreatorID
	}
	return ChannelCreation{
		Name:      trimmed,
		Private:   private,
		CreatorID: creatorID,
		MemberIDs: dedupeExcept(memberIDs, creatorID),
	}, nil
}

// dedupeExcept はidsから重複要素と、excludeに一致する要素（作成者自身が誤ってメンションされた場合）
// を除去した新しいスライスを返す。呼び出し順（メンションされた順）は維持する
func dedupeExcept(ids []string, exclude string) []string {
	seen := map[string]bool{exclude: true}
	var result []string
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}
