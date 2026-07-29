package notifychannel

import (
	"context"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

type fakeRepository struct {
	items map[string]notifychannel.NotifyChannel // key: guildID+purpose
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{items: map[string]notifychannel.NotifyChannel{}}
}

func key(guildID string, purpose notifychannel.Purpose) string {
	return guildID + "|" + string(purpose)
}

func (f *fakeRepository) Set(ctx context.Context, nc notifychannel.NotifyChannel) error {
	f.items[key(nc.GuildID, nc.Purpose)] = nc
	return nil
}

func (f *fakeRepository) Find(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error) {
	nc, ok := f.items[key(guildID, purpose)]
	return nc, ok, nil
}

func TestUseCase_SetAndGet(t *testing.T) {
	tests := []struct {
		name          string
		setGuildID    string
		setPurpose    notifychannel.Purpose
		setChannelID  string
		getGuildID    string
		getPurpose    notifychannel.Purpose
		wantOK        bool
		wantChannelID string
	}{
		{
			name:          "正常系: 設定した通知先チャンネルを同じギルド・用途で取得できる",
			setGuildID:    "g1",
			setPurpose:    notifychannel.PurposeEmoji,
			setChannelID:  "c1",
			getGuildID:    "g1",
			getPurpose:    notifychannel.PurposeEmoji,
			wantOK:        true,
			wantChannelID: "c1",
		},
		{
			name:         "異常系: 別の用途で設定した場合は取得できない",
			setGuildID:   "g1",
			setPurpose:   notifychannel.PurposeEmoji,
			setChannelID: "c1",
			getGuildID:   "g1",
			getPurpose:   notifychannel.Purpose("other"),
			wantOK:       false,
		},
		{
			name:         "異常系: 未設定のギルドは取得できない",
			setGuildID:   "g1",
			setPurpose:   notifychannel.PurposeEmoji,
			setChannelID: "c1",
			getGuildID:   "g2",
			getPurpose:   notifychannel.PurposeEmoji,
			wantOK:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			u := New(repo)
			if err := u.Set(context.Background(), tt.setGuildID, tt.setPurpose, tt.setChannelID); err != nil {
				t.Fatalf("Set() unexpected error = %v", err)
			}
			got, ok, err := u.Get(context.Background(), tt.getGuildID, tt.getPurpose)
			if err != nil {
				t.Fatalf("Get() unexpected error = %v", err)
			}
			if ok != tt.wantOK {
				t.Fatalf("Get() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got.ChannelID != tt.wantChannelID {
				t.Errorf("Get() channelID = %q, want %q", got.ChannelID, tt.wantChannelID)
			}
		})
	}
}
