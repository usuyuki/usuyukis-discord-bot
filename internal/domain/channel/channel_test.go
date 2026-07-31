package channel

import (
	"strings"
	"testing"
)

func TestNewName(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "正常系: 前後の空白を除いた文字列がそのままNameになる", raw: "  general-chat  ", want: "general-chat", wantErr: nil},
		{name: "異常系: 空文字を入れるとErrEmptyNameになる", raw: "", wantErr: ErrEmptyName},
		{name: "異常系: 空白のみを入れるとErrEmptyNameになる", raw: "   ", wantErr: ErrEmptyName},
		{name: "異常系: 101文字を入れると上限を超えるのでErrNameTooLongになる", raw: strings.Repeat("a", 101), wantErr: ErrNameTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewName(tt.raw)
			if err != tt.wantErr {
				t.Fatalf("NewName(%q) error = %v, want %v", tt.raw, err, tt.wantErr)
			}
			if tt.wantErr == nil && got.String() != tt.want {
				t.Errorf("NewName(%q).String() = %q, want %q", tt.raw, got.String(), tt.want)
			}
		})
	}
}

func TestNewName_MaxLengthIsAllowed(t *testing.T) {
	raw := strings.Repeat("a", 100)
	got, err := NewName(raw)
	if err != nil {
		t.Fatalf("NewName() unexpected error = %v", err)
	}
	if got.String() != raw {
		t.Errorf("NewName().String() = %q, want %q", got.String(), raw)
	}
}
