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

func TestDetectsMentionsBot(t *testing.T) {
	tests := []struct {
		name           string
		mentionUserIDs []string
		content        string
		botID          string
		botNames       []string
		want           bool
	}{
		{
			name:           "正常系: 構造化メンションのIDが一致すれば真",
			mentionUserIDs: []string{"bot1"},
			content:        "<@bot1> help",
			botID:          "bot1",
			botNames:       []string{"usuyuki"},
			want:           true,
		},
		{
			name:           "異常系: メンションもbotName一致もなければ偽",
			mentionUserIDs: []string{"user1"},
			content:        "hello",
			botID:          "bot1",
			botNames:       []string{"usuyuki"},
			want:           false,
		},
		{
			name:           "正常系: 構造化メンションが失われた地の文の@botName表記でも真になる（コピペ救済）",
			mentionUserIDs: nil,
			content:        "@usuyuki help",
			botID:          "bot1",
			botNames:       []string{"usuyuki"},
			want:           true,
		},
		{
			name:           "正常系: 地の文@botName表記は大文字小文字を無視する",
			mentionUserIDs: nil,
			content:        "@Usuyuki help",
			botID:          "bot1",
			botNames:       []string{"usuyuki"},
			want:           true,
		},
		{
			name:           "異常系: botNameが文中（先頭以外）に現れても偽",
			mentionUserIDs: nil,
			content:        "help @usuyuki",
			botID:          "bot1",
			botNames:       []string{"usuyuki"},
			want:           false,
		},
		{
			name:           "異常系: botNamesが空（State.User未同期）なら地の文判定は行わない",
			mentionUserIDs: nil,
			content:        "@usuyuki help",
			botID:          "bot1",
			botNames:       nil,
			want:           false,
		},
		{
			name:           "正常系: 先頭がGlobalName（表示名）と一致すれば真になる（Username以外の表示名でのコピペ救済）",
			mentionUserIDs: nil,
			content:        "@うすゆき help",
			botID:          "bot1",
			botNames:       []string{"usuyuki", "うすゆき"},
			want:           true,
		},
		{
			name:           "正常系: 全角＠のプレフィックスでも救済判定が成立する（日本語IME/モバイル入力対策）",
			mentionUserIDs: nil,
			content:        "＠usuyuki help",
			botID:          "bot1",
			botNames:       []string{"usuyuki"},
			want:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectsMentionsBot(tt.mentionUserIDs, tt.content, tt.botID, tt.botNames); got != tt.want {
				t.Errorf("detectsMentionsBot(%v, %q, %q, %v) = %v, want %v", tt.mentionUserIDs, tt.content, tt.botID, tt.botNames, got, tt.want)
			}
		})
	}
}

func TestBotNameCandidates(t *testing.T) {
	tests := []struct {
		name string
		user *discordgo.User
		want []string
	}{
		{
			name: "正常系: UsernameとGlobalNameが異なれば両方を候補にする",
			user: &discordgo.User{Username: "usuyuki_bot", GlobalName: "うすゆき"},
			want: []string{"usuyuki_bot", "うすゆき"},
		},
		{
			name: "正常系: GlobalName未設定ならUsernameのみを候補にする",
			user: &discordgo.User{Username: "usuyuki_bot", GlobalName: ""},
			want: []string{"usuyuki_bot"},
		},
		{
			name: "異常系: GlobalNameがUsernameと同一なら重複させない",
			user: &discordgo.User{Username: "usuyuki_bot", GlobalName: "usuyuki_bot"},
			want: []string{"usuyuki_bot"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := botNameCandidates(tt.user)
			if len(got) != len(tt.want) {
				t.Fatalf("botNameCandidates() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("botNameCandidates() = %v, want %v", got, tt.want)
				}
			}
		})
	}
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
