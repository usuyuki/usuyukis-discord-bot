package channelcreate

import (
	"context"
	"errors"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channelcreate"
)

// fakeChannelCreator はテスト用のChannelCreatorフェイク実装
type fakeChannelCreator struct {
	gotGuildID  string
	gotCreation channelcreate.ChannelCreation
	channelID   string
	err         error
}

func (f *fakeChannelCreator) CreateChannel(ctx context.Context, guildID string, creation channelcreate.ChannelCreation) (string, error) {
	f.gotGuildID = guildID
	f.gotCreation = creation
	return f.channelID, f.err
}

func TestUseCase_Create(t *testing.T) {
	tests := []struct {
		name          string
		guildID       string
		channelName   string
		private       bool
		creatorID     string
		memberIDs     []string
		creatorErr    error
		wantChannelID string
		wantErr       bool
	}{
		{
			name:          "正常系: 公開チャンネルを作成できる",
			guildID:       "g1",
			channelName:   "general-2",
			creatorID:     "u1",
			wantChannelID: "c1",
		},
		{
			name:          "正常系: プライベートチャンネルを作成できる",
			guildID:       "g1",
			channelName:   "secret",
			private:       true,
			creatorID:     "u1",
			memberIDs:     []string{"u2"},
			wantChannelID: "c2",
		},
		{
			name:        "異常系: 名前が空文字だとport呼び出し前に検証エラーになる",
			guildID:     "g1",
			channelName: "",
			creatorID:   "u1",
			wantErr:     true,
		},
		{
			name:        "異常系: ChannelCreatorがエラーを返すとCreateもエラーを返す",
			guildID:     "g1",
			channelName: "general",
			creatorID:   "u1",
			creatorErr:  errors.New("boom"),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := &fakeChannelCreator{channelID: tt.wantChannelID, err: tt.creatorErr}
			u := New(creator)

			got, err := u.Create(context.Background(), tt.guildID, tt.channelName, tt.private, tt.creatorID, tt.memberIDs)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Create() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Create() unexpected error = %v", err)
			}
			if got != tt.wantChannelID {
				t.Errorf("Create() = %q, want %q", got, tt.wantChannelID)
			}
			if creator.gotGuildID != tt.guildID {
				t.Errorf("CreateChannel() guildID = %q, want %q", creator.gotGuildID, tt.guildID)
			}
			if creator.gotCreation.Private != tt.private {
				t.Errorf("CreateChannel() creation.Private = %v, want %v", creator.gotCreation.Private, tt.private)
			}
		})
	}
}
