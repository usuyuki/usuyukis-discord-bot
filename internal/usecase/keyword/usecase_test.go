package keyword

import (
	"context"
	"errors"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
)

// fakeRepository はテスト用のインメモリRepository実装。
// guildID -> word -> responses の形で保持し、AddResponse/RemoveResponse/RemoveKeywordの
// 実際の永続化層（postgres）に近い挙動（応答の積み増し・個別削除）を再現する
type fakeRepository struct {
	items map[string]map[string][]string // guildID -> word -> responses
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{items: map[string]map[string][]string{}}
}

func (f *fakeRepository) AddResponse(ctx context.Context, guildID, word, response string) error {
	if f.items[guildID] == nil {
		f.items[guildID] = map[string][]string{}
	}
	f.items[guildID][word] = append(f.items[guildID][word], response)
	return nil
}

func (f *fakeRepository) RemoveResponse(ctx context.Context, guildID, word, response string) error {
	responses := f.items[guildID][word]
	kept := responses[:0]
	for _, r := range responses {
		if r != response {
			kept = append(kept, r)
		}
	}
	if len(kept) == 0 {
		delete(f.items[guildID], word)
		return nil
	}
	f.items[guildID][word] = kept
	return nil
}

func (f *fakeRepository) RemoveKeyword(ctx context.Context, guildID, word string) error {
	delete(f.items[guildID], word)
	return nil
}

func (f *fakeRepository) ReplaceResponses(ctx context.Context, guildID, word string, responses []string) error {
	if f.items[guildID] == nil {
		f.items[guildID] = map[string][]string{}
	}
	f.items[guildID][word] = responses
	return nil
}

func (f *fakeRepository) FindByGuild(ctx context.Context, guildID string) ([]keyword.Keyword, error) {
	var result []keyword.Keyword
	for word, responses := range f.items[guildID] {
		k, err := keyword.New(0, guildID, word, responses)
		if err != nil {
			return nil, err
		}
		result = append(result, k)
	}
	return result, nil
}

type errorRepository struct{ err error }

func (e *errorRepository) AddResponse(ctx context.Context, guildID, word, response string) error {
	return e.err
}
func (e *errorRepository) RemoveResponse(ctx context.Context, guildID, word, response string) error {
	return e.err
}
func (e *errorRepository) RemoveKeyword(ctx context.Context, guildID, word string) error {
	return e.err
}
func (e *errorRepository) ReplaceResponses(ctx context.Context, guildID, word string, responses []string) error {
	return e.err
}
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
		{name: "異常系: responseが空文字だとドメインバリデーションエラーになる", guildID: "g1", word: "ぬるぽ", response: "", wantErr: true},
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

func TestUseCase_Register_Accumulates(t *testing.T) {
	repo := newFakeRepository()
	u := New(repo)
	ctx := context.Background()

	if err := u.Register(ctx, "g1", "ぬるぽ", "ガッ"); err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}
	if err := u.Register(ctx, "g1", "ぬるぽ", "ｶﾞｯ"); err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	keywords, err := u.List(ctx, "g1")
	if err != nil {
		t.Fatalf("List() unexpected error = %v", err)
	}
	if len(keywords) != 1 {
		t.Fatalf("List() returned %d keywords, want 1", len(keywords))
	}
	if len(keywords[0].Responses) != 2 {
		t.Fatalf("List() keyword has %d responses, want 2 (accumulated)", len(keywords[0].Responses))
	}
}

func TestUseCase_RemoveResponse(t *testing.T) {
	repo := newFakeRepository()
	u := New(repo)
	ctx := context.Background()
	if err := u.Register(ctx, "g1", "ぬるぽ", "ガッ"); err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}
	if err := u.Register(ctx, "g1", "ぬるぽ", "ｶﾞｯ"); err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	if err := u.RemoveResponse(ctx, "g1", "ぬるぽ", "ガッ"); err != nil {
		t.Fatalf("RemoveResponse() unexpected error = %v", err)
	}

	keywords, err := u.List(ctx, "g1")
	if err != nil {
		t.Fatalf("List() unexpected error = %v", err)
	}
	if len(keywords) != 1 || len(keywords[0].Responses) != 1 || keywords[0].Responses[0] != "ｶﾞｯ" {
		t.Fatalf("List() = %+v, want single keyword with response %q remaining", keywords, "ｶﾞｯ")
	}
}

func TestUseCase_RemoveResponse_LastOneRemovesKeyword(t *testing.T) {
	repo := newFakeRepository()
	u := New(repo)
	ctx := context.Background()
	if err := u.Register(ctx, "g1", "ぬるぽ", "ガッ"); err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	if err := u.RemoveResponse(ctx, "g1", "ぬるぽ", "ガッ"); err != nil {
		t.Fatalf("RemoveResponse() unexpected error = %v", err)
	}

	keywords, err := u.List(ctx, "g1")
	if err != nil {
		t.Fatalf("List() unexpected error = %v", err)
	}
	if len(keywords) != 0 {
		t.Fatalf("List() = %+v, want no keywords remaining after last response removed", keywords)
	}
}

func TestUseCase_RemoveKeyword(t *testing.T) {
	repo := newFakeRepository()
	u := New(repo)
	ctx := context.Background()
	if err := u.Register(ctx, "g1", "ぬるぽ", "ガッ"); err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}
	if err := u.Register(ctx, "g1", "ぬるぽ", "ｶﾞｯ"); err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	if err := u.RemoveKeyword(ctx, "g1", "ぬるぽ"); err != nil {
		t.Fatalf("RemoveKeyword() unexpected error = %v", err)
	}

	keywords, err := u.List(ctx, "g1")
	if err != nil {
		t.Fatalf("List() unexpected error = %v", err)
	}
	if len(keywords) != 0 {
		t.Fatalf("List() = %+v, want no keywords remaining after RemoveKeyword", keywords)
	}
}

func TestUseCase_SetResponses(t *testing.T) {
	tests := []struct {
		name      string
		guildID   string
		word      string
		responses []string
		wantErr   bool
	}{
		{name: "正常系: 複数応答で丸ごと置き換えられる", guildID: "g1", word: "ぬるぽ", responses: []string{"ガッ", "ｶﾞｯ"}, wantErr: false},
		{name: "異常系: responsesが空だとドメインバリデーションエラーになる", guildID: "g1", word: "ぬるぽ", responses: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			u := New(repo)
			err := u.SetResponses(context.Background(), tt.guildID, tt.word, tt.responses)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SetResponses() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			keywords, err := u.List(context.Background(), tt.guildID)
			if err != nil {
				t.Fatalf("List() unexpected error = %v", err)
			}
			if len(keywords) != 1 || len(keywords[0].Responses) != len(tt.responses) {
				t.Fatalf("List() = %+v, want single keyword with %d responses", keywords, len(tt.responses))
			}
		})
	}
}

func TestUseCase_Match(t *testing.T) {
	tests := []struct {
		name          string
		registered    map[string][]string // word -> responses
		registerGuild string
		guildID       string
		messageBody   string
		wantOK        bool
		wantWord      string
	}{
		{
			name:          "正常系: 登録済みキーワードが本文に含まれていれば一致する",
			registered:    map[string][]string{"ぬるぽ": {"ガッ"}},
			registerGuild: "g1",
			guildID:       "g1",
			messageBody:   "さっきぬるぽ食らった",
			wantOK:        true,
			wantWord:      "ぬるぽ",
		},
		{
			name:          "異常系: 一致するキーワードがなければokはfalse",
			registered:    map[string][]string{"ぬるぽ": {"ガッ"}},
			registerGuild: "g1",
			guildID:       "g1",
			messageBody:   "今日はいい天気",
			wantOK:        false,
		},
		{
			name:          "異常系: 別ギルドに登録されたキーワードとは一致しない",
			registered:    map[string][]string{"ぬるぽ": {"ガッ"}},
			registerGuild: "g2",
			guildID:       "g1",
			messageBody:   "ぬるぽ",
			wantOK:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			for word, responses := range tt.registered {
				repo.items[tt.registerGuild] = map[string][]string{word: responses}
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

func TestUseCase_Match_ReturnsKeywordWithAllResponses(t *testing.T) {
	repo := newFakeRepository()
	repo.items["g1"] = map[string][]string{"ぬるぽ": {"ガッ", "ｶﾞｯ"}}
	u := New(repo)
	got, ok, err := u.Match(context.Background(), "g1", "ぬるぽ")
	if err != nil {
		t.Fatalf("Match() unexpected error = %v", err)
	}
	if !ok {
		t.Fatal("Match() ok = false, want true")
	}
	if len(got.Responses) != 2 {
		t.Fatalf("Match() Responses = %v, want 2 responses so caller can pick randomly", got.Responses)
	}
}
