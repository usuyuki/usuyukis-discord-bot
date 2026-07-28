package keyword

import (
	"context"
	"errors"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
)

// fakeRepository はテスト用のインメモリRepository実装
type fakeRepository struct {
	items map[string][]keyword.Keyword // guildID -> keywords
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{items: map[string][]keyword.Keyword{}}
}

func (f *fakeRepository) Save(ctx context.Context, k keyword.Keyword) error {
	f.items[k.GuildID] = append(f.items[k.GuildID], k)
	return nil
}

func (f *fakeRepository) Delete(ctx context.Context, guildID, word string) error {
	kept := f.items[guildID][:0]
	for _, k := range f.items[guildID] {
		if k.Word != word {
			kept = append(kept, k)
		}
	}
	f.items[guildID] = kept
	return nil
}

func (f *fakeRepository) FindByGuild(ctx context.Context, guildID string) ([]keyword.Keyword, error) {
	return f.items[guildID], nil
}

type errorRepository struct{ err error }

func (e *errorRepository) Save(ctx context.Context, k keyword.Keyword) error      { return e.err }
func (e *errorRepository) Delete(ctx context.Context, guildID, word string) error { return e.err }
func (e *errorRepository) FindByGuild(ctx context.Context, guildID string) ([]keyword.Keyword, error) {
	return nil, e.err
}

func TestUseCase_Register(t *testing.T) {
	tests := []struct {
		name     string
		guildID  string
		word     string
		response string
		wantErr  bool
	}{
		{name: "正常系: 有効な値なら登録できる", guildID: "g1", word: "ぬるぽ", response: "ガッ", wantErr: false},
		{name: "異常系: wordが空文字だとドメインバリデーションエラーになる", guildID: "g1", word: "", response: "ガッ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			u := New(repo)
			err := u.Register(context.Background(), tt.guildID, tt.word, tt.response)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUseCase_Match(t *testing.T) {
	tests := []struct {
		name        string
		registered  []keyword.Keyword
		guildID     string
		messageBody string
		wantOK      bool
		wantWord    string
	}{
		{
			name: "正常系: 登録済みキーワードが本文に含まれていれば一致する",
			registered: []keyword.Keyword{
				mustNewKeyword(t, "g1", "ぬるぽ", "ガッ"),
			},
			guildID:     "g1",
			messageBody: "さっきぬるぽ食らった",
			wantOK:      true,
			wantWord:    "ぬるぽ",
		},
		{
			name: "異常系: 一致するキーワードがなければokはfalse",
			registered: []keyword.Keyword{
				mustNewKeyword(t, "g1", "ぬるぽ", "ガッ"),
			},
			guildID:     "g1",
			messageBody: "今日はいい天気",
			wantOK:      false,
		},
		{
			name: "異常系: 別ギルドに登録されたキーワードとは一致しない",
			registered: []keyword.Keyword{
				mustNewKeyword(t, "g2", "ぬるぽ", "ガッ"),
			},
			guildID:     "g1",
			messageBody: "ぬるぽ",
			wantOK:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			for _, k := range tt.registered {
				repo.items[k.GuildID] = append(repo.items[k.GuildID], k)
			}
			u := New(repo)
			got, ok, err := u.Match(context.Background(), tt.guildID, tt.messageBody)
			if err != nil {
				t.Fatalf("Match() unexpected error = %v", err)
			}
			if ok != tt.wantOK {
				t.Fatalf("Match() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got.Word != tt.wantWord {
				t.Errorf("Match() word = %q, want %q", got.Word, tt.wantWord)
			}
		})
	}
}

func TestUseCase_Match_RepositoryError(t *testing.T) {
	wantErr := errors.New("db down")
	u := New(&errorRepository{err: wantErr})
	_, _, err := u.Match(context.Background(), "g1", "ぬるぽ")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Match() error = %v, want %v", err, wantErr)
	}
}

func mustNewKeyword(t *testing.T, guildID, word, response string) keyword.Keyword {
	t.Helper()
	k, err := keyword.New(0, guildID, word, response)
	if err != nil {
		t.Fatalf("keyword.New() unexpected error = %v", err)
	}
	return k
}
