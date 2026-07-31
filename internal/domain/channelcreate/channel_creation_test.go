package channelcreate

import (
	"reflect"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		rawName   string
		private   bool
		creatorID string
		memberIDs []string
		want      ChannelCreation
		wantErr   error
	}{
		{
			name:      "正常系: 公開チャンネルとして生成できる",
			rawName:   "general-2",
			private:   false,
			creatorID: "u1",
			want:      ChannelCreation{Name: "general-2", Private: false, CreatorID: "u1"},
		},
		{
			name:      "正常系: 前後の空白は除去される",
			rawName:   "  general-2  ",
			creatorID: "u1",
			want:      ChannelCreation{Name: "general-2", CreatorID: "u1"},
		},
		{
			name:      "正常系: プライベートチャンネルはメンバーIDの重複を除去して保持する",
			rawName:   "secret",
			private:   true,
			creatorID: "u1",
			memberIDs: []string{"u2", "u3", "u2"},
			want:      ChannelCreation{Name: "secret", Private: true, CreatorID: "u1", MemberIDs: []string{"u2", "u3"}},
		},
		{
			name:      "正常系: 作成者自身がメンバーIDに含まれていても除去される",
			rawName:   "secret",
			private:   true,
			creatorID: "u1",
			memberIDs: []string{"u1", "u2"},
			want:      ChannelCreation{Name: "secret", Private: true, CreatorID: "u1", MemberIDs: []string{"u2"}},
		},
		{
			name:      "異常系: 名前が空文字だとErrEmptyNameになる",
			rawName:   "",
			creatorID: "u1",
			wantErr:   ErrEmptyName,
		},
		{
			name:      "異常系: 名前が空白のみだとErrEmptyNameになる",
			rawName:   "   ",
			creatorID: "u1",
			wantErr:   ErrEmptyName,
		},
		{
			name:      "異常系: 名前が101文字を超えるとErrNameTooLongになる",
			rawName:   strings.Repeat("a", 101),
			creatorID: "u1",
			wantErr:   ErrNameTooLong,
		},
		{
			name:    "異常系: creatorIDが空文字だとErrEmptyCreatorIDになる",
			rawName: "general",
			wantErr: ErrEmptyCreatorID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.rawName, tt.private, tt.creatorID, tt.memberIDs)
			if err != tt.wantErr {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestNew_NameLengthBoundary(t *testing.T) {
	name100 := strings.Repeat("a", 100)
	if _, err := New(name100, false, "u1", nil); err != nil {
		t.Errorf("New() with 100-char name unexpected error = %v", err)
	}
}
