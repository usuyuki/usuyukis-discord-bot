package admin

import (
	"context"
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/keyword"
	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/notifychannel"
)

//go:embed templates/*.html
var templateFS embed.FS

// KeywordUseCase はadminサーバーが必要とするキーワード操作
type KeywordUseCase interface {
	Register(ctx context.Context, guildID, word, response string) error
	RemoveKeyword(ctx context.Context, guildID, word string) error
	SetResponses(ctx context.Context, guildID, word string, responses []string) error
	List(ctx context.Context, guildID string) ([]keyword.Keyword, error)
}

// NotifyChannelUseCase はadminサーバーが必要とする通知先チャンネル操作
type NotifyChannelUseCase interface {
	Set(ctx context.Context, guildID string, purpose notifychannel.Purpose, channelID string) error
	Get(ctx context.Context, guildID string, purpose notifychannel.Purpose) (notifychannel.NotifyChannel, bool, error)
}

// Server は認証なし・localhost限定を前提とする管理画面のHTTPサーバー
type Server struct {
	guilds   GuildDirectory
	keywords KeywordUseCase
	notify   NotifyChannelUseCase
	pages    map[string]*template.Template // page名(例: "guilds.html") -> layout+content結合済みテンプレート
}

// NewServer はServerを生成する。layout.htmlとページ別htmlの組み合わせごとに
// 独立したtemplate.Templateを構築する（html/templateはdefineブロック名が
// ファイル間で重複すると後勝ちで上書きされてしまうため、まとめてParseFSしない）
func NewServer(guilds GuildDirectory, keywords KeywordUseCase, notify NotifyChannelUseCase) (*Server, error) {
	pageFiles := []string{"guilds.html", "guild_detail.html"}
	pages := make(map[string]*template.Template, len(pageFiles))
	for _, page := range pageFiles {
		tmpl, err := template.ParseFS(templateFS, "templates/layout.html", "templates/"+page)
		if err != nil {
			return nil, err
		}
		pages[page] = tmpl
	}
	return &Server{guilds: guilds, keywords: keywords, notify: notify, pages: pages}, nil
}

// Handler は管理画面のルーティングを構成したhttp.Handlerを返す
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleGuildList)
	mux.HandleFunc("GET /guilds/{guildID}", s.handleGuildDetail)
	mux.HandleFunc("POST /guilds/{guildID}/keywords", s.handleKeywordCreate)
	mux.HandleFunc("POST /guilds/{guildID}/keywords/update", s.handleKeywordUpdate)
	mux.HandleFunc("POST /guilds/{guildID}/keywords/delete", s.handleKeywordDelete)
	mux.HandleFunc("POST /guilds/{guildID}/notify-channels", s.handleNotifyChannelSet)
	return mux
}

func (s *Server) handleGuildList(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title  string
		Guilds []GuildInfo
	}{
		Title:  "ギルド一覧",
		Guilds: s.guilds.ListGuilds(),
	}
	s.render(w, "guilds.html", data)
}

func (s *Server) handleGuildDetail(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("guildID")
	ctx := r.Context()

	keywords, err := s.keywords.List(ctx, guildID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	channels, err := s.guilds.ListTextChannels(guildID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	haikuChannelName := channelNameForPurpose(ctx, s, guildID, notifychannel.PurposeHaiku, channels)
	emojiChannelName := channelNameForPurpose(ctx, s, guildID, notifychannel.PurposeEmoji, channels)

	data := struct {
		Title            string
		GuildID          string
		GuildName        string
		Keywords         []keyword.Keyword
		Channels         []ChannelInfo
		HaikuChannelName string
		EmojiChannelName string
	}{
		Title:            "ギルド詳細",
		GuildID:          guildID,
		GuildName:        s.guilds.GuildName(guildID),
		Keywords:         keywords,
		Channels:         channels,
		HaikuChannelName: haikuChannelName,
		EmojiChannelName: emojiChannelName,
	}
	s.render(w, "guild_detail.html", data)
}

func channelNameForPurpose(ctx context.Context, s *Server, guildID string, purpose notifychannel.Purpose, channels []ChannelInfo) string {
	nc, ok, err := s.notify.Get(ctx, guildID, purpose)
	if err != nil || !ok {
		return "未設定"
	}
	for _, c := range channels {
		if c.ID == nc.ChannelID {
			return c.Name
		}
	}
	return nc.ChannelID
}

func (s *Server) handleKeywordCreate(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("guildID")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	word := r.PostForm.Get("word")
	response := r.PostForm.Get("response")

	if err := s.keywords.Register(r.Context(), guildID, word, response); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/guilds/"+guildID, http.StatusSeeOther)
}

func (s *Server) handleKeywordUpdate(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("guildID")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	word := r.PostForm.Get("word")
	// 応答一覧は改行区切りのテキストエリアで受け取り、空行は除外して丸ごと置き換える
	responses := make([]string, 0)
	for _, line := range strings.Split(r.PostForm.Get("responses"), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			responses = append(responses, trimmed)
		}
	}

	if err := s.keywords.SetResponses(r.Context(), guildID, word, responses); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/guilds/"+guildID, http.StatusSeeOther)
}

func (s *Server) handleKeywordDelete(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("guildID")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	word := r.PostForm.Get("word")

	if err := s.keywords.RemoveKeyword(r.Context(), guildID, word); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/guilds/"+guildID, http.StatusSeeOther)
}

func (s *Server) handleNotifyChannelSet(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("guildID")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	purpose := notifychannel.Purpose(r.PostForm.Get("purpose"))
	channelID := r.PostForm.Get("channel_id")

	if err := s.notify.Set(r.Context(), guildID, purpose, channelID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/guilds/"+guildID, http.StatusSeeOther)
}

func (s *Server) render(w http.ResponseWriter, page string, data any) {
	tmpl, ok := s.pages[page]
	if !ok {
		http.Error(w, "admin: unknown page "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
