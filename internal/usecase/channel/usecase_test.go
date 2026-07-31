package channel

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channel"
)

// fakeCreator はテスト用のCreatorフェイク実装。CreatePrivateChannelに渡された引数を記録する
type fakeCreator struct {
	gotGuildID   string
	gotName      string
	gotCreatorID string
	channelID    string
	err          error
}

func (f *fakeCreator) CreatePrivateChannel(ctx context.Context, guildID, name, creatorUserID string) (string, error) {
	f.gotGuildID = guildID
	f.gotName = name
	f.gotCreatorID = creatorUserID
	return f.channelID, f.err
}

func TestUseCase_Create(t *testing.T) {
	t.Run("正常系: 有効な名前ならCreatorへ委譲し作成されたチャンネルIDを返す", func(t *testing.T) {
		creator := &fakeCreator{channelID: "ch1"}
		u := New(creator)

		got, err := u.Create(context.Background(), "g1", "  雑談  ", "user1")
		if err != nil {
			t.Fatalf("Create() unexpected error = %v", err)
		}
		if got != "ch1" {
			t.Errorf("Create() = %q, want %q", got, "ch1")
		}
		if creator.gotGuildID != "g1" || creator.gotName != "雑談" || creator.gotCreatorID != "user1" {
			t.Errorf("Creator received guildID=%q name=%q creatorID=%q, want g1/雑談/user1", creator.gotGuildID, creator.gotName, creator.gotCreatorID)
		}
	})

	t.Run("異常系: 空白のみの名前だとCreatorを呼ばずErrEmptyNameになる", func(t *testing.T) {
		creator := &fakeCreator{}
		u := New(creator)

		_, err := u.Create(context.Background(), "g1", "   ", "user1")
		if !errors.Is(err, channel.ErrEmptyName) {
			t.Fatalf("Create() error = %v, want %v", err, channel.ErrEmptyName)
		}
		if creator.gotGuildID != "" {
			t.Errorf("Creator should not be called on validation error, but was called with guildID=%q", creator.gotGuildID)
		}
	})

	t.Run("異常系: 長すぎる名前だとCreatorを呼ばずErrNameTooLongになる", func(t *testing.T) {
		creator := &fakeCreator{}
		u := New(creator)

		_, err := u.Create(context.Background(), "g1", strings.Repeat("a", 101), "user1")
		if !errors.Is(err, channel.ErrNameTooLong) {
			t.Fatalf("Create() error = %v, want %v", err, channel.ErrNameTooLong)
		}
	})

	t.Run("異常系: Creatorがエラーを返すとCreateもエラーを返す", func(t *testing.T) {
		wantErr := errors.New("boom")
		creator := &fakeCreator{err: wantErr}
		u := New(creator)

		_, err := u.Create(context.Background(), "g1", "雑談", "user1")
		if !errors.Is(err, wantErr) {
			t.Fatalf("Create() error = %v, want %v", err, wantErr)
		}
	})
}
