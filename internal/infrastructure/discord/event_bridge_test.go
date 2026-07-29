package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

// newTestSession はState上にGuild/Role/Memberを登録済みのSessionを組み立てる。
// StateFullなので個々のAdd系メソッドが単体で反映される
func newTestSession(t *testing.T, guildID, ownerID string, roles []*discordgo.Role, member *discordgo.Member) *discordgo.Session {
	t.Helper()
	state := discordgo.NewState()
	guild := &discordgo.Guild{ID: guildID, OwnerID: ownerID, Roles: roles}
	if err := state.GuildAdd(guild); err != nil {
		t.Fatalf("GuildAdd failed: %v", err)
	}
	if member != nil {
		member.GuildID = guildID
		if err := state.MemberAdd(member); err != nil {
			t.Fatalf("MemberAdd failed: %v", err)
		}
	}
	return &discordgo.Session{State: state}
}

func TestDefaultAdminPermissionChecker(t *testing.T) {
	const guildID = "g1"
	everyoneRole := &discordgo.Role{ID: guildID, Permissions: 0}
	adminRole := &discordgo.Role{ID: "role-admin", Permissions: discordgo.PermissionAdministrator}
	memberRole := &discordgo.Role{ID: "role-member", Permissions: discordgo.PermissionViewChannel}

	tests := []struct {
		name    string
		userID  string
		ownerID string
		roles   []*discordgo.Role
		member  *discordgo.Member
		want    bool
	}{
		{
			name:    "正常系: guildのOwnerIDが対象userIDと一致するとownerとして管理者扱いになる",
			userID:  "owner1",
			ownerID: "owner1",
			roles:   []*discordgo.Role{everyoneRole},
			member:  &discordgo.Member{User: &discordgo.User{ID: "owner1"}, Roles: nil},
			want:    true,
		},
		{
			name:    "正常系: Administrator権限を持つロールを付与されたメンバーは管理者と判定される",
			userID:  "user1",
			ownerID: "owner1",
			roles:   []*discordgo.Role{everyoneRole, adminRole},
			member:  &discordgo.Member{User: &discordgo.User{ID: "user1"}, Roles: []string{"role-admin"}},
			want:    true,
		},
		{
			name:    "正常系: Administrator権限を持たないロールのみのメンバーは管理者と判定されない",
			userID:  "user2",
			ownerID: "owner1",
			roles:   []*discordgo.Role{everyoneRole, memberRole},
			member:  &discordgo.Member{User: &discordgo.User{ID: "user2"}, Roles: []string{"role-member"}},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSession(t, guildID, tt.ownerID, tt.roles, tt.member)
			got, err := DefaultAdminPermissionChecker(s, guildID, tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("DefaultAdminPermissionChecker() = %v, want %v", got, tt.want)
			}
		})
	}
}
