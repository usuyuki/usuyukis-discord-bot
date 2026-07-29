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

func TestKeyword_RandomResponse(t *testing.T) {
	fixedNow := time.Date(2026, 7, 28, 14, 32, 10, 0, time.UTC)

	t.Run("正常系: 応答が1件のみなら常にその応答が返る", func(t *testing.T) {
		k, err := New(1, "g1", "ぬるぽ", []string{"ガッ"})
		if err != nil {
			t.Fatalf("New() unexpected error = %v", err)
		}
		for range 10 {
			if got := k.RandomResponse(fixedNow); got != "ガッ" {
				t.Fatalf("RandomResponse() = %q, want %q", got, "ガッ")
			}
		}
	})

	t.Run("正常系: 複数応答のいずれかが返り、登録外の応答は返らない", func(t *testing.T) {
		responses := []string{"ガッ", "ｶﾞｯ", "がっ"}
		k, err := New(1, "g1", "ぬるぽ", responses)
		if err != nil {
			t.Fatalf("New() unexpected error = %v", err)
		}
		seen := map[string]bool{}
		for range 100 {
			got := k.RandomResponse(fixedNow)
			found := false
			for _, r := range responses {
				if r == got {
					found = true
				}
			}
			if !found {
				t.Fatalf("RandomResponse() = %q, want one of %v", got, responses)
			}
			seen[got] = true
		}
		if len(seen) < 2 {
			t.Fatalf("RandomResponse() over 100 trials only produced %v, expected more variety", seen)
		}
	})

	t.Run("正常系: 応答に{$now}が含まれていれば現在時刻のJST表記に展開される", func(t *testing.T) {
		k, err := New(1, "g1", "今何時", []string{"今は{$now}だよ"})
		if err != nil {
			t.Fatalf("New() unexpected error = %v", err)
		}
		want := "今は2026-07-28 23:32:10だよ"
		if got := k.RandomResponse(fixedNow); got != want {
			t.Fatalf("RandomResponse() = %q, want %q", got, want)
		}
	})

	t.Run("正常系: {$now}が含まれていなければそのまま返る", func(t *testing.T) {
		k, err := New(1, "g1", "ぬるぽ", []string{"ガッ"})
		if err != nil {
			t.Fatalf("New() unexpected error = %v", err)
		}
		if got := k.RandomResponse(fixedNow); got != "ガッ" {
			t.Fatalf("RandomResponse() = %q, want %q", got, "ガッ")
		}
	})
}
