package discordbot

import (
	"context"
	"errors"
	"testing"

	channelUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/channel"
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

func TestChannelHandler_HandleChannelCreate(t *testing.T) {
	t.Run("正常系: プライベートチャンネルならRestrictorが呼ばれる", func(t *testing.T) {
		restrictor := &fakeRestrictor{}
		h := NewChannelHandler(channelUC.New(restrictor))

		ev := IncomingChannelCreate{GuildID: "g1", ChannelID: "c1", CreatorID: "user1", IsPrivate: true}
		if err := h.HandleChannelCreate(context.Background(), ev); err != nil {
			t.Fatalf("HandleChannelCreate() unexpected error = %v", err)
		}
		if !restrictor.called {
			t.Fatal("HandleChannelCreate() should call Restrictor for a private channel")
		}
		if restrictor.gotGuildID != "g1" || restrictor.gotChannelID != "c1" || restrictor.gotCreatorID != "user1" {
			t.Errorf("Restrictor received guildID=%q channelID=%q creatorID=%q, want g1/c1/user1", restrictor.gotGuildID, restrictor.gotChannelID, restrictor.gotCreatorID)
		}
	})

	t.Run("異常系: プライベートでなければRestrictorは呼ばれない", func(t *testing.T) {
		restrictor := &fakeRestrictor{}
		h := NewChannelHandler(channelUC.New(restrictor))

		ev := IncomingChannelCreate{GuildID: "g1", ChannelID: "c1", CreatorID: "user1", IsPrivate: false}
		if err := h.HandleChannelCreate(context.Background(), ev); err != nil {
			t.Fatalf("HandleChannelCreate() unexpected error = %v", err)
		}
		if restrictor.called {
			t.Error("HandleChannelCreate() should not call Restrictor for a non-private channel")
		}
	})

	t.Run("異常系: Restrictorのエラーがそのまま返る", func(t *testing.T) {
		wantErr := errors.New("discord api boom")
		restrictor := &fakeRestrictor{err: wantErr}
		h := NewChannelHandler(channelUC.New(restrictor))

		ev := IncomingChannelCreate{GuildID: "g1", ChannelID: "c1", CreatorID: "user1", IsPrivate: true}
		err := h.HandleChannelCreate(context.Background(), ev)
		if !errors.Is(err, wantErr) {
			t.Fatalf("HandleChannelCreate() error = %v, want %v", err, wantErr)
		}
	})
}
