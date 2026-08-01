package channel

import "testing"

const (
	testPermViewChannel    int64 = 1 << 10
	testPermManageChannels int64 = 1 << 4
	testPermAdministrator  int64 = 1 << 3
)

func TestIsPrivate(t *testing.T) {
	tests := []struct {
		name       string
		guildID    string
		overwrites []Overwrite
		want       bool
	}{
		{
			name:       "正常系: @everyoneロールへViewChannel拒否があればプライベート判定",
			guildID:    "g1",
			overwrites: []Overwrite{{RoleID: "g1", Deny: testPermViewChannel}},
			want:       true,
		},
		{
			name:       "正常系: @everyoneロールへの拒否がViewChannel以外のみなら非プライベート判定",
			guildID:    "g1",
			overwrites: []Overwrite{{RoleID: "g1", Deny: testPermManageChannels}},
			want:       false,
		},
		{
			name:       "異常系: @everyone以外のロールへのViewChannel拒否だけでは非プライベート判定",
			guildID:    "g1",
			overwrites: []Overwrite{{RoleID: "role0", Deny: testPermViewChannel}},
			want:       false,
		},
		{
			name:       "異常系: オーバーライドが空なら非プライベート判定",
			guildID:    "g1",
			overwrites: nil,
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPrivate(tt.guildID, tt.overwrites); got != tt.want {
				t.Errorf("IsPrivate(%q, %v) = %v, want %v", tt.guildID, tt.overwrites, got, tt.want)
			}
		})
	}
}

func TestIsChannelManagerRole(t *testing.T) {
	tests := []struct {
		name        string
		permissions int64
		want        bool
	}{
		{
			name:        "正常系: ManageChannelsのみ持つロールは対象になる",
			permissions: testPermManageChannels,
			want:        true,
		},
		{
			name:        "異常系: ManageChannelsとAdministratorを両方持つロールはAdministratorが常にオーバーライドをバイパスするため対象外",
			permissions: testPermManageChannels | testPermAdministrator,
			want:        false,
		},
		{
			name:        "異常系: ManageChannelsを持たないロールは対象外",
			permissions: 0,
			want:        false,
		},
		{
			name:        "異常系: Administratorのみ持つロールは対象外",
			permissions: testPermAdministrator,
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsChannelManagerRole(tt.permissions); got != tt.want {
				t.Errorf("IsChannelManagerRole(%d) = %v, want %v", tt.permissions, got, tt.want)
			}
		})
	}
}
