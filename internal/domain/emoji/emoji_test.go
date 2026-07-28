package emoji

import "testing"

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		emojiName string
		id        string
		animated  bool
		wantErr   error
	}{
		{name: "正常系: 名前とIDが埋まっていれば生成できる", emojiName: "youkoso_nya", id: "123", animated: false, wantErr: nil},
		{name: "正常系: アニメーション絵文字も生成できる", emojiName: "youkoso_nya", id: "123", animated: true, wantErr: nil},
		{name: "異常系: nameが空文字を入れるとErrEmptyNameエラーになる", emojiName: "", id: "123", animated: false, wantErr: ErrEmptyName},
		{name: "異常系: idが空文字を入れるとErrEmptyIDエラーになる", emojiName: "youkoso_nya", id: "", animated: false, wantErr: ErrEmptyID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.emojiName, tt.id, tt.animated)
			if err != tt.wantErr {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmoji_Tag(t *testing.T) {
	tests := []struct {
		name      string
		emojiName string
		id        string
		animated  bool
		want      string
	}{
		{name: "正常系: 通常絵文字は<:name:id>形式になる", emojiName: "youkoso_nya", id: "123", animated: false, want: "<:youkoso_nya:123>"},
		{name: "正常系: アニメーション絵文字は<a:name:id>形式になる", emojiName: "youkoso_nya", id: "123", animated: true, want: "<a:youkoso_nya:123>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := New(tt.emojiName, tt.id, tt.animated)
			if err != nil {
				t.Fatalf("New() unexpected error = %v", err)
			}
			if got := e.Tag(); got != tt.want {
				t.Errorf("Tag() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEmoji_Name(t *testing.T) {
	e, err := New("youkoso_nya", "123", false)
	if err != nil {
		t.Fatalf("New() unexpected error = %v", err)
	}
	if got := e.Name(); got != "youkoso_nya" {
		t.Errorf("Name() = %q, want %q", got, "youkoso_nya")
	}
}
