package channel

import "testing"

func TestIsApproved(t *testing.T) {
	tests := []struct {
		name              string
		approverCount     int
		requiredApprovals int
		want              bool
	}{
		{
			name:              "正常系: 承認者数が必要数と一致すると可決される",
			approverCount:     2,
			requiredApprovals: 2,
			want:              true,
		},
		{
			name:              "正常系: 承認者数が必要数を上回っても可決される",
			approverCount:     3,
			requiredApprovals: 2,
			want:              true,
		},
		{
			name:              "異常系: 承認者数が必要数を下回ると可決されない",
			approverCount:     1,
			requiredApprovals: 2,
			want:              false,
		},
		{
			name:              "異常系: 承認者が0人だと可決されない",
			approverCount:     0,
			requiredApprovals: 2,
			want:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsApproved(tt.approverCount, tt.requiredApprovals); got != tt.want {
				t.Errorf("IsApproved(%d, %d) = %v, want %v", tt.approverCount, tt.requiredApprovals, got, tt.want)
			}
		})
	}
}
