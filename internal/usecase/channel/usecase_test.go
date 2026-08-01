package channel

import (
	"context"
	"errors"
	"testing"
)

// fakeRestrictor はテスト用のRestrictorフェイク実装。呼び出されたかと引数を記録する
type fakeRestrictor struct {
	called       bool
	gotGuildID   string
	gotChannelID string
	gotCreatorID string
	err          error
}

func (f *fakeRestrictor) RestrictToCreatorAndAdmins(ctx context.Context, guildID, channelID, creatorUserID string) error {
	f.called = true
	f.gotGuildID = guildID
	f.gotChannelID = channelID
	f.gotCreatorID = creatorUserID
	return f.err
}

func TestUseCase_ProtectIfPrivate(t *testing.T) {
	t.Run("正常系: プライベートチャンネルならRestrictorへ委譲する", func(t *testing.T) {
		restrictor := &fakeRestrictor{}
		u := New(restrictor)

		if err := u.ProtectIfPrivate(context.Background(), "g1", "ch1", "user1", true); err != nil {
			t.Fatalf("ProtectIfPrivate() unexpected error = %v", err)
		}
		if !restrictor.called {
			t.Fatal("ProtectIfPrivate() should call Restrictor when isPrivate is true")
		}
		if restrictor.gotGuildID != "g1" || restrictor.gotChannelID != "ch1" || restrictor.gotCreatorID != "user1" {
			t.Errorf("Restrictor received guildID=%q channelID=%q creatorID=%q, want g1/ch1/user1", restrictor.gotGuildID, restrictor.gotChannelID, restrictor.gotCreatorID)
		}
	})

	t.Run("異常系: プライベートでなければRestrictorを呼ばない", func(t *testing.T) {
		restrictor := &fakeRestrictor{}
		u := New(restrictor)

		if err := u.ProtectIfPrivate(context.Background(), "g1", "ch1", "user1", false); err != nil {
			t.Fatalf("ProtectIfPrivate() unexpected error = %v", err)
		}
		if restrictor.called {
			t.Error("ProtectIfPrivate() should not call Restrictor when isPrivate is false")
		}
	})

	t.Run("異常系: Restrictorがエラーを返すとProtectIfPrivateもエラーを返す", func(t *testing.T) {
		wantErr := errors.New("boom")
		restrictor := &fakeRestrictor{err: wantErr}
		u := New(restrictor)

		err := u.ProtectIfPrivate(context.Background(), "g1", "ch1", "user1", true)
		if !errors.Is(err, wantErr) {
			t.Fatalf("ProtectIfPrivate() error = %v, want %v", err, wantErr)
		}
	})
}
