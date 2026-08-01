package channel

import "testing"

func TestNewRequiredApprovals(t *testing.T) {
	tests := []struct {
		name    string
		raw     int
		wantErr bool
		want    int
	}{
		{
			name: "正常系: 1以上の値はそのまま登録できる",
			raw:  2,
			want: 2,
		},
		{
			name: "正常系: 最小値1も登録できる",
			raw:  1,
			want: 1,
		},
		{
			name:    "異常系: 0を入れると誰も承認せず作成される矛盾が起きるのでエラーになる",
			raw:     0,
			wantErr: true,
		},
		{
			name:    "異常系: 負数を入れると意味を成さないのでエラーになる",
			raw:     -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRequiredApprovals(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewRequiredApprovals(%d) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Int() != tt.want {
				t.Errorf("NewRequiredApprovals(%d).Int() = %d, want %d", tt.raw, got.Int(), tt.want)
			}
		})
	}
}

func TestDefaultRequiredApprovals(t *testing.T) {
	if got := DefaultRequiredApprovals().Int(); got != 2 {
		t.Errorf("DefaultRequiredApprovals().Int() = %d, want 2", got)
	}
}
