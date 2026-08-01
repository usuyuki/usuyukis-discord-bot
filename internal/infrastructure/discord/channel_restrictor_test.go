package discord

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestExistingRoleAllowBits(t *testing.T) {
	tests := []struct {
		name       string
		overwrites []*discordgo.PermissionOverwrite
		want       map[string]int64
	}{
		{
			name: "正常系: ロール単位のオーバーライドのAllowビットのみ抽出される",
			overwrites: []*discordgo.PermissionOverwrite{
				{ID: "role1", Type: discordgo.PermissionOverwriteTypeRole, Allow: discordgo.PermissionSendMessages},
				{ID: "user1", Type: discordgo.PermissionOverwriteTypeMember, Allow: discordgo.PermissionViewChannel},
			},
			want: map[string]int64{"role1": discordgo.PermissionSendMessages},
		},
		{
			name:       "異常系: オーバーライドが空だとAllowビットが1件も無いので空マップになる",
			overwrites: nil,
			want:       map[string]int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSession(t, "g1", "owner1", nil, nil)
			ch := &discordgo.Channel{ID: "ch1", GuildID: "g1", PermissionOverwrites: tt.overwrites}
			if err := s.State.ChannelAdd(ch); err != nil {
				t.Fatalf("ChannelAdd failed: %v", err)
			}

			got := existingRoleAllowBits(s, "ch1")
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("existingRoleAllowBits() = %v, want %v", got, tt.want)
			}
		})
	}
}
