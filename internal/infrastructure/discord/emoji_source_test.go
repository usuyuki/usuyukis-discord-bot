package discord

import (
	"context"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestEmojiSource_ListEmojiTags(t *testing.T) {
	tests := []struct {
		name    string
		guildID string
		guild   *discordgo.Guild
		queryID string
		want    []string
		wantErr bool
	}{
		{
			name:    "正常系: ギルドに登録済みの絵文字がタグ文字列のリストとして返る",
			guildID: "g1",
			guild: &discordgo.Guild{
				ID: "g1",
				Emojis: []*discordgo.Emoji{
					{ID: "1", Name: "a"},
					{ID: "2", Name: "b", Animated: true},
				},
			},
			queryID: "g1",
			want:    []string{"<:a:1>", "<a:b:2>"},
		},
		{
			name:    "正常系: 絵文字が未登録のギルドは空スライスを返す",
			guildID: "g1",
			guild:   &discordgo.Guild{ID: "g1", Emojis: nil},
			queryID: "g1",
			want:    []string{},
		},
		{
			name:    "異常系: Stateに存在しないguildIDを指定するとエラーになる",
			guildID: "g1",
			guild:   &discordgo.Guild{ID: "g1"},
			queryID: "does-not-exist",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := discordgo.NewState()
			if err := state.GuildAdd(tt.guild); err != nil {
				t.Fatalf("GuildAdd failed: %v", err)
			}
			session := &discordgo.Session{State: state}
			s := NewEmojiSource(session)

			got, err := s.ListEmojiTags(context.Background(), tt.queryID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ListEmojiTags() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ListEmojiTags() unexpected error = %v", err)
			}
			if !equalStringSlice(got, tt.want) {
				t.Errorf("ListEmojiTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
