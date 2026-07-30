package keyword

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		guildID   string
		word      string
		responses []string
		wantErr   error
	}{
		{name: "正常系: 全項目が埋まっていれば生成できる", guildID: "g1", word: "ぬるぽ", responses: []string{"ガッ"}, wantErr: nil},
		{name: "正常系: 応答が複数あっても生成できる", guildID: "g1", word: "ぬるぽ", responses: []string{"ガッ", "ｶﾞｯ"}, wantErr: nil},
		{name: "異常系: guildIDが空文字だとErrEmptyGuildIDになる", guildID: "", word: "ぬるぽ", responses: []string{"ガッ"}, wantErr: ErrEmptyGuildID},
		{name: "異常系: wordが空文字だとErrEmptyKeywordになる", guildID: "g1", word: "", responses: []string{"ガッ"}, wantErr: ErrEmptyKeyword},
		{name: "異常系: responsesが空だとErrEmptyResponseになる", guildID: "g1", word: "ぬるぽ", responses: nil, wantErr: ErrEmptyResponse},
		{name: "異常系: responsesに空文字が含まれるとErrEmptyResponseになる", guildID: "g1", word: "ぬるぽ", responses: []string{"ガッ", ""}, wantErr: ErrEmptyResponse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(1, tt.guildID, tt.word, tt.responses)
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
			k, err := New(1, "g1", tt.word, []string{"ガッ"})
			if err != nil {
				t.Fatalf("New() unexpected error = %v", err)
			}
			if got := k.Matches(tt.messageBody); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyword_ResponseAt(t *testing.T) {
	fixedNow := time.Date(2026, 7, 28, 14, 32, 10, 0, time.UTC)

	t.Run("正常系: 指定インデックスの応答が返る", func(t *testing.T) {
		responses := []string{"ガッ", "ｶﾞｯ", "がっ"}
		k, err := New(1, "g1", "ぬるぽ", responses)
		if err != nil {
			t.Fatalf("New() unexpected error = %v", err)
		}
		for i, want := range responses {
			if got := k.ResponseAt(i, fixedNow); got != want {
				t.Fatalf("ResponseAt(%d) = %q, want %q", i, got, want)
			}
		}
	})

	t.Run("正常系: 応答に{$now}が含まれていれば現在時刻のJST表記に展開される", func(t *testing.T) {
		k, err := New(1, "g1", "今何時", []string{"今は{$now}だよ"})
		if err != nil {
			t.Fatalf("New() unexpected error = %v", err)
		}
		want := "今は2026-07-28 23:32:10だよ"
		if got := k.ResponseAt(0, fixedNow); got != want {
			t.Fatalf("ResponseAt() = %q, want %q", got, want)
		}
	})

	t.Run("正常系: {$now}が含まれていなければそのまま返る", func(t *testing.T) {
		k, err := New(1, "g1", "ぬるぽ", []string{"ガッ"})
		if err != nil {
			t.Fatalf("New() unexpected error = %v", err)
		}
		if got := k.ResponseAt(0, fixedNow); got != "ガッ" {
			t.Fatalf("ResponseAt() = %q, want %q", got, "ガッ")
		}
	})
}
