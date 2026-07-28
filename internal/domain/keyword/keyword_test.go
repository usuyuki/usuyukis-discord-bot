package keyword

import "testing"

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		guildID  string
		word     string
		response string
		wantErr  error
	}{
		{name: "正常系: 全項目が埋まっていれば生成できる", guildID: "g1", word: "ぬるぽ", response: "ガッ", wantErr: nil},
		{name: "異常系: guildIDが空文字だとErrEmptyGuildIDになる", guildID: "", word: "ぬるぽ", response: "ガッ", wantErr: ErrEmptyGuildID},
		{name: "異常系: wordが空文字だとErrEmptyKeywordになる", guildID: "g1", word: "", response: "ガッ", wantErr: ErrEmptyKeyword},
		{name: "異常系: responseが空文字だとErrEmptyResponseになる", guildID: "g1", word: "ぬるぽ", response: "", wantErr: ErrEmptyResponse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(1, tt.guildID, tt.word, tt.response)
			if err != tt.wantErr {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestKeyword_Matches(t *testing.T) {
	tests := []struct {
		name        string
		word        string
		messageBody string
		want        bool
	}{
		{name: "正常系: キーワードが本文に部分一致で含まれていればtrue", word: "ぬるぽ", messageBody: "さっきぬるぽ食らった", want: true},
		{name: "正常系: 本文とキーワードが完全一致してもtrue", word: "ぬるぽ", messageBody: "ぬるぽ", want: true},
		{name: "異常系: キーワードが本文に含まれないのでfalseになる", word: "ぬるぽ", messageBody: "今日はいい天気ですね", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, err := New(1, "g1", tt.word, "ガッ")
			if err != nil {
				t.Fatalf("New() unexpected error = %v", err)
			}
			if got := k.Matches(tt.messageBody); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}
