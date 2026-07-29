package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

type fakeGuildDirectory struct {
	guilds   []GuildInfo
	channels map[string][]ChannelInfo
}

func (f *fakeGuildDirectory) ListGuilds() []GuildInfo { return f.guilds }
func (f *fakeGuildDirectory) ListTextChannels(guildID string) ([]ChannelInfo, error) {
	return f.channels[guildID], nil
}
func (f *fakeGuildDirectory) GuildName(guildID string) string {
	for _, g := range f.guilds {
		if g.ID == guildID {
			return g.Name
		}
	}
	return guildID
}

type fakeKeywordUseCase struct {
	items map[string][]keyword.Keyword
}

func newFakeKeywordUseCase() *fakeKeywordUseCase {
	return &fakeKeywordUseCase{items: map[string][]keyword.Keyword{}}
}

func (f *fakeKeywordUseCase) Register(ctx context.Context, guildID, word, response string) error {
	for i, k := range f.items[guildID] {
		if k.Word == word {
			updated, err := keyword.New(k.ID, guildID, word, append(k.Responses, response))
			if err != nil {
				return err
			}
			f.items[guildID][i] = updated
			return nil
		}
	}
	k, err := keyword.New(0, guildID, word, []string{response})
	if err != nil {
		return err
	}
	f.items[guildID] = append(f.items[guildID], k)
	return nil
}

func (f *fakeKeywordUseCase) RemoveKeyword(ctx context.Context, guildID, word string) error {
	kept := f.items[guildID][:0]
	for _, k := range f.items[guildID] {
		if k.Word != word {
			kept = append(kept, k)
		}
	}
	f.items[guildID] = kept
	return nil
}

func (f *fakeKeywordUseCase) SetResponses(ctx context.Context, guildID, word string, responses []string) error {
	for i, k := range f.items[guildID] {
		if k.Word == word {
			updated, err := keyword.New(k.ID, guildID, word, responses)
			if err != nil {
				return err
			}
			f.items[guildID][i] = updated
			return nil
		}
	}
	k, err := keyword.New(0, guildID, word, responses)
	if err != nil {
		return err
	}
	f.items[guildID] = append(f.items[guildID], k)
	return nil
}

func (f *fakeKeywordUseCase) List(ctx context.Context, guildID string) ([]keyword.Keyword, error) {
	return f.items[guildID], nil
}

type fakeNotifyChannelUseCase struct {
	items map[string]notifychannel.NotifyChannel
}

func newFakeNotifyChannelUseCase() *fakeNotifyChannelUseCase {
	return &fakeNotifyChannelUseCase{items: map[string]notifychannel.NotifyChannel{}}
}

func (f *fakeNotifyChannelUseCase) Set(ctx context.Context, guildID string, purpose notifychannel.Purpose, channelID string) error {
	nc, err := notifychannel.New(guildID, purpose, channelID)
	if err != nil {
		return err
	}
	f.items[guildID+"|"+string(purpose)] = nc
	return nil
}

func (f *fakeNotifyChannelUseCase) Get(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error) {
	nc, ok := f.items[guildID+"|"+string(purpose)]
	return nc, ok, nil
}

func newTestServer(t *testing.T) (*Server, *fakeKeywordUseCase, *fakeNotifyChannelUseCase) {
	t.Helper()
	guilds := &fakeGuildDirectory{
		guilds:   []GuildInfo{{ID: "g1", Name: "テストギルド"}},
		channels: map[string][]ChannelInfo{"g1": {{ID: "c1", Name: "general"}}},
	}
	kw := newFakeKeywordUseCase()
	nc := newFakeNotifyChannelUseCase()
	s, err := NewServer(guilds, kw, nc)
	if err != nil {
		t.Fatalf("NewServer() unexpected error = %v", err)
	}
	return s, kw, nc
}

func TestServer_GuildList(t *testing.T) {
	s, _, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "テストギルド") {
		t.Errorf("GET / body does not contain guild name: %s", rec.Body.String())
	}
}

func TestServer_KeywordCreateAndDelete(t *testing.T) {
	s, kw, _ := newTestServer(t)

	t.Run("正常系: フォームPOSTでキーワードが登録される", func(t *testing.T) {
		form := url.Values{"word": {"ぬるぽ"}, "response": {"ガッ"}}
		req := httptest.NewRequest(http.MethodPost, "/guilds/g1/keywords", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		s.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusSeeOther {
			t.Fatalf("POST keywords status = %d, want %d", rec.Code, http.StatusSeeOther)
		}
		if len(kw.items["g1"]) != 1 || kw.items["g1"][0].Word != "ぬるぽ" {
			t.Errorf("keyword was not registered: %v", kw.items["g1"])
		}
	})

	t.Run("正常系: フォームPOSTで応答一覧を改行区切りで丸ごと更新できる", func(t *testing.T) {
		form := url.Values{"word": {"ぬるぽ"}, "responses": {"ガッ\nｶﾞｯ\n"}}
		req := httptest.NewRequest(http.MethodPost, "/guilds/g1/keywords/update", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		s.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusSeeOther {
			t.Fatalf("POST keywords/update status = %d, want %d", rec.Code, http.StatusSeeOther)
		}
		if len(kw.items["g1"]) != 1 || len(kw.items["g1"][0].Responses) != 2 {
			t.Errorf("responses were not replaced: %v", kw.items["g1"])
		}
	})

	t.Run("正常系: フォームPOSTでキーワードが削除される", func(t *testing.T) {
		form := url.Values{"word": {"ぬるぽ"}}
		req := httptest.NewRequest(http.MethodPost, "/guilds/g1/keywords/delete", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		s.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusSeeOther {
			t.Fatalf("POST keywords/delete status = %d, want %d", rec.Code, http.StatusSeeOther)
		}
		if len(kw.items["g1"]) != 0 {
			t.Errorf("keyword was not deleted: %v", kw.items["g1"])
		}
	})
}

func TestServer_NotifyChannelSet(t *testing.T) {
	s, _, nc := newTestServer(t)

	form := url.Values{"purpose": {"emoji"}, "channel_id": {"c1"}}
	req := httptest.NewRequest(http.MethodPost, "/guilds/g1/notify-channels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST notify-channels status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	got, ok, _ := nc.Get(context.Background(), "g1", notifychannel.PurposeEmoji)
	if !ok || got.ChannelID != "c1" {
		t.Errorf("notify channel was not set correctly: ok=%v got=%v", ok, got)
	}
}

func TestServer_GuildDetail(t *testing.T) {
	s, kw, _ := newTestServer(t)
	_ = kw.Register(context.Background(), "g1", "ぬるぽ", "ガッ")

	req := httptest.NewRequest(http.MethodGet, "/guilds/g1", nil)
	rec := httptest.NewRecorder()

	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /guilds/g1 status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "ぬるぽ") || !strings.Contains(body, "ガッ") {
		t.Errorf("GET /guilds/g1 body does not contain registered keyword: %s", body)
	}
}
