package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	mediadb "github.com/goritskimihail/mudro/internal/media"
	"github.com/goritskimihail/mudro/internal/posts"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	pool             *pgxpool.Pool
	postsSvc         *posts.Service
	authHandlers     *AuthHandlers
	adminHandlers    *AdminHandlers
	tgVisiblePostIDs []string
}

func NewServer(pool *pgxpool.Pool, postsSvc *posts.Service, handlers ...any) *Server {
	s := &Server{
		pool:     pool,
		postsSvc: postsSvc,
	}
	for _, handler := range handlers {
		switch v := handler.(type) {
		case *AuthHandlers:
			s.authHandlers = v
		case *AdminHandlers:
			s.adminHandlers = v
		}
	}
	return s
}

func (s *Server) Router() http.Handler {
	mediaRoot := strings.TrimSpace(os.Getenv("MEDIA_ROOT"))
	if mediaRoot == "" {
		mediaRoot = filepath.Join(config.RepoRoot(), "data", "nu")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/feed", http.StatusTemporaryRedirect)
	})
	mux.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(mediaRoot))))
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/posts", s.handlePosts)
	mux.HandleFunc("/api/front", s.handleFront)
	mux.HandleFunc("/api/orchestration/status", s.handleOrchestrationStatus)
	mux.HandleFunc("/feed", s.handleFeed)

	mux.HandleFunc("/api/casino", s.handleCasinoIndex)

	// Auth routes.
	if s.authHandlers != nil {
		mux.HandleFunc("/api/casino/", s.authHandlers.AuthMiddleware(s.handleCasinoProxy))
		mux.HandleFunc("/api/auth/register", s.authHandlers.HandleRegister)
		mux.HandleFunc("/api/auth/login", s.authHandlers.HandleLogin)
		mux.HandleFunc("/api/auth/logout", s.authHandlers.HandleLogout)
		mux.HandleFunc("/api/auth/me", s.authHandlers.AuthMiddleware(s.authHandlers.HandleMe))

		if s.adminHandlers != nil {
			mux.HandleFunc("/api/admin/users", s.authHandlers.AuthAdminMiddleware(s.adminHandlers.HandleGetUsers))
			mux.HandleFunc("/api/admin/stats", s.authHandlers.AuthAdminMiddleware(s.adminHandlers.HandleGetStats))
		}
	}

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With, X-User-ID, X-User-Name, X-User-Role, X-User-Email")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleCasinoIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":   "ok",
		"service":  "casino",
		"base_url": config.CasinoServiceURL(),
	})
}

func (s *Server) handleCasinoProxy(w http.ResponseWriter, r *http.Request) {
	target, err := url.Parse(config.CasinoServiceURL())
	if err != nil {
		log.Printf("casino target: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("casino proxy: %v", err)
		http.Error(rw, "casino service unavailable", http.StatusBadGateway)
	}
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/casino")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		req.Host = target.Host
		req.Header.Del("Authorization")
		req.Header.Del("Cookie")

		if user := UserFromContext(req.Context()); user != nil {
			req.Header.Set("X-User-ID", strconv.FormatInt(user.ID, 10))
			req.Header.Set("X-User-Name", user.Username)
			req.Header.Set("X-User-Role", user.Role)
			if user.Email != nil {
				req.Header.Set("X-User-Email", strings.TrimSpace(*user.Email))
			} else {
				req.Header.Del("X-User-Email")
			}
		}
	}
	proxy.ServeHTTP(w, r)
}

func (s *Server) handlePosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := parseLimit(r.URL.Query().Get("limit"))
	page, err := parsePage(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "invalid page: "+err.Error(), http.StatusBadRequest)
		return
	}

	cursorTS, cursorID, err := parseCursor(r.URL.Query().Get("before_ts"), r.URL.Query().Get("before_id"))
	if err != nil {
		http.Error(w, "invalid cursor: "+err.Error(), http.StatusBadRequest)
		return
	}
	source, err := parseSource(r.URL.Query().Get("source"))
	if err != nil {
		http.Error(w, "invalid source: "+err.Error(), http.StatusBadRequest)
		return
	}
	sortOrder, err := parseSort(r.URL.Query().Get("sort"))
	if err != nil {
		http.Error(w, "invalid sort: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))

	items, next, err := s.postsSvc.LoadPosts(ctx, cursorTS, cursorID, page, limit, source, posts.SortOrder(sortOrder), query)
	if err != nil {
		log.Printf("loadPosts: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := postsResponse{
		Page:       page,
		Limit:      limit,
		Items:      items,
		NextCursor: next,
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}

func (s *Server) handleFront(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit := parseLimit(r.URL.Query().Get("limit"))
	source, err := parseSource(r.URL.Query().Get("source"))
	if err != nil {
		http.Error(w, "invalid source: "+err.Error(), http.StatusBadRequest)
		return
	}
	sortOrder, err := parseSort(r.URL.Query().Get("sort"))
	if err != nil {
		http.Error(w, "invalid sort: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))

	items, next, err := s.postsSvc.LoadPosts(ctx, nil, nil, nil, limit, source, posts.SortOrder(sortOrder), query)
	if err != nil {
		log.Printf("loadPosts(front): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	totalPosts, err := s.postsSvc.CountVisiblePosts(ctx)
	if err != nil {
		log.Printf("front count posts: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var lastSync *time.Time
	if err := s.pool.QueryRow(ctx, `select max(updated_at) from posts`).Scan(&lastSync); err != nil {
		log.Printf("front max(updated_at): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sourceStats, err := s.postsSvc.LoadSourceStats(ctx)
	if err != nil {
		log.Printf("front source stats: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := frontResponse{
		Meta: frontMeta{
			TotalPosts: totalPosts,
			LastSyncAt: lastSync,
			Sources:    sourceStats,
		},
		Feed: postsResponse{
			Limit:      limit,
			Items:      items,
			NextCursor: next,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}

func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, _, err := parseCursor(r.URL.Query().Get("before_ts"), r.URL.Query().Get("before_id")); err != nil {
		http.Error(w, "invalid cursor: "+err.Error(), http.StatusBadRequest)
		return
	}
	limit := parseLimit(r.URL.Query().Get("limit"))
	page, err := parsePage(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "invalid page: "+err.Error(), http.StatusBadRequest)
		return
	}
	if page == nil {
		defaultPage := 1
		page = &defaultPage
	}
	source, err := parseSource(r.URL.Query().Get("source"))
	if err != nil {
		http.Error(w, "invalid source: "+err.Error(), http.StatusBadRequest)
		return
	}
	sortOrder, err := parseSort(r.URL.Query().Get("sort"))
	if err != nil {
		http.Error(w, "invalid sort: "+err.Error(), http.StatusBadRequest)
		return
	}

	q := r.URL.Query().Get("q")
	items, _, err := s.postsSvc.LoadPosts(ctx, nil, nil, page, limit, source, posts.SortOrder(sortOrder), q)
	if err != nil {
		log.Printf("loadPosts(feed): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var totalPosts int64
	if err := s.countVisiblePosts(ctx, &totalPosts); err != nil {
		log.Printf("feed count posts: %v", err)
	}
	sourceStats, err := s.loadSourceStats(ctx)
	if err != nil {
		log.Printf("feed source stats: %v", err)
	}
	vkTotal, tgTotal := sourceTotals(sourceStats)

	data := feedPageData{
		Limit:      limit,
		Page:       *page,
		Source:     source,
		SortOrder:  sortOrder,
		AllURL:     buildFeedURL(limit, 1, "", sortOrder),
		VKURL:      buildFeedURL(limit, 1, "vk", sortOrder),
		TGURL:      buildFeedURL(limit, 1, "tg", sortOrder),
		NewestURL:  buildFeedURL(limit, 1, source, "desc"),
		OldestURL:  buildFeedURL(limit, 1, source, "asc"),
		Items:      make([]feedItem, 0, len(items)),
		SourceName: sourceLabel(source),
		TotalPosts: totalPosts,
		VKTotal:    vkTotal,
		TGTotal:    tgTotal,
	}

	for _, p := range items {
		originalURL := buildOriginalPostURL(p.Source, p.SourcePostID)
		commentsURL := ""
		if p.CommentsCount != nil && *p.CommentsCount > 0 && originalURL != "" {
			commentsURL = originalURL
		}
		data.Items = append(data.Items, feedItem{
			ID:            p.ID,
			Source:        p.Source,
			SourcePostID:  p.SourcePostID,
			PublishedAt:   p.PublishedAt.Format("2006-01-02 15:04:05"),
			Text:          compactText(p.Text),
			LikesCount:    p.LikesCount,
			ViewsCount:    p.ViewsCount,
			CommentsCount: p.CommentsCount,
			OriginalURL:   originalURL,
			CommentsURL:   commentsURL,
			Reactions:     buildFeedReactions(p.Reactions),
			Comments:      buildFeedComments(p.Comments),
			Media:         parseMediaItems(p.Media),
		})
	}

	if len(items) == limit {
		data.NextURL = buildFeedURL(limit, *page+1, source, sortOrder)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := feedPageTmpl.Execute(w, data); err != nil {
		log.Printf("template feed: %v", err)
	}
}

func compactText(s *string) string {
	if s == nil {
		return ""
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return ""
	}
	return t
}

func buildFeedReactions(reactions map[string]int) []feedReaction {
	if len(reactions) == 0 {
		return nil
	}
	keys := make([]string, 0, len(reactions))
	for k, v := range reactions {
		if v > 0 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		vi := reactions[keys[i]]
		vj := reactions[keys[j]]
		if vi == vj {
			return keys[i] < keys[j]
		}
		return vi > vj
	})
	out := make([]feedReaction, 0, len(keys))
	for _, k := range keys {
		out = append(out, feedReaction{
			Label: reactionLabel(k),
			Count: reactions[k],
			Raw:   k,
		})
	}
	return out
}

func reactionLabel(raw string) string {
	switch {
	case strings.HasPrefix(raw, "emoji:"):
		label := strings.TrimSpace(strings.TrimPrefix(raw, "emoji:"))
		if label != "" {
			return label
		}
	case strings.HasPrefix(raw, "custom:"):
		return "✨"
	case strings.HasPrefix(raw, "unknown:"):
		return "?"
	}
	if raw == "" {
		return "?"
	}
	return raw
}

func buildFeedComments(comments []posts.Comment) []feedComment {
	if len(comments) == 0 {
		return nil
	}
	out := make([]feedComment, 0, len(comments))
	for _, c := range comments {
		out = append(out, feedComment{
			SourceCommentID: c.SourceCommentID,
			ParentCommentID: c.ParentCommentID,
			AuthorName:      c.AuthorName,
			PublishedAt:     c.PublishedAt,
			Text:            c.Text,
			Reactions:       buildFeedReactions(c.Reactions),
			Media:           buildFeedMedia(c.Media),
		})
	}
	return out
}

func buildFeedMedia(media json.RawMessage) []feedMediaItem {
	if len(media) == 0 {
		return nil
	}
	items := posts.ParseMediaItems(media)
	out := make([]feedMediaItem, 0, len(items))
	for _, item := range items {
		out = append(out, feedMediaItem{
			Kind:       item.Kind,
			URL:        item.URL,
			PreviewURL: item.PreviewURL,
			Title:      item.Title,
			Width:      item.Width,
			Height:     item.Height,
			Position:   item.Position,
			IsImage:    item.Kind == "photo" || item.Kind == "gif" || item.Kind == "image",
			IsAudio:    item.Kind == "audio",
			IsVideo:    item.Kind == "video",
			IsDocument: item.Kind == "doc",
			IsLink:     item.Kind == "link",
		})
	}
	return out
}

func sourceTotals(stats []posts.SourceStat) (vkTotal, tgTotal int64) {
	for _, st := range stats {
		switch st.Source {
		case "vk":
			vkTotal = st.Posts
		case "tg":
			tgTotal = st.Posts
		}
	}
	return vkTotal, tgTotal
}

type postsResponse struct {
	Page       *int          `json:"page,omitempty"`
	Limit      int           `json:"limit"`
	Items      []posts.Post  `json:"items"`
	NextCursor *posts.Cursor `json:"next_cursor,omitempty"`
}

type frontResponse struct {
	Meta frontMeta     `json:"meta"`
	Feed postsResponse `json:"feed"`
}

type frontMeta struct {
	TotalPosts int64              `json:"total_posts"`
	LastSyncAt *time.Time         `json:"last_sync_at,omitempty"`
	Sources    []posts.SourceStat `json:"sources"`
}

type feedPageData struct {
	Limit      int
	Page       int
	Source     string
	SourceName string
	SortOrder  string
	TotalPosts int64
	VKTotal    int64
	TGTotal    int64
	AllURL     string
	VKURL      string
	TGURL      string
	NewestURL  string
	OldestURL  string
	Items      []feedItem
	NextURL    string
}

type feedItem struct {
	ID            int64
	Source        string
	SourcePostID  string
	PublishedAt   string
	Text          string
	LikesCount    int
	ViewsCount    *int
	CommentsCount *int
	OriginalURL   string
	CommentsURL   string
	Reactions     []feedReaction
	Comments      []feedComment
	Media         []feedMediaItem
}

type feedReaction struct {
	Label string
	Count int
	Raw   string
}

type feedComment struct {
	SourceCommentID string          `json:"source_comment_id"`
	ParentCommentID string          `json:"parent_comment_id,omitempty"`
	AuthorName      string          `json:"author_name"`
	PublishedAt     string          `json:"published_at"`
	Text            string          `json:"text"`
	Reactions       []feedReaction  `json:"reactions,omitempty"`
	Media           []feedMediaItem `json:"media,omitempty"`
}

type feedMediaItem struct {
	Kind       string `json:"kind"`
	URL        string `json:"url"`
	PreviewURL string `json:"preview_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Position   int    `json:"position,omitempty"`
	IsImage    bool   `json:"is_image,omitempty"`
	IsAudio    bool   `json:"is_audio,omitempty"`
	IsVideo    bool   `json:"is_video,omitempty"`
	IsDocument bool   `json:"is_document,omitempty"`
	IsLink     bool   `json:"is_link,omitempty"`
}

var feedPageTmpl = template.Must(template.New("feed").Parse(`<!doctype html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Mudro Feed</title>
  <style>
    :root {
      --bg: #f5f7fb;
      --card: #ffffff;
      --text: #1f2937;
      --muted: #64748b;
      --line: #e2e8f0;
      --accent: #0f766e;
    }
    body {
      margin: 0;
      background: radial-gradient(circle at top left, #e0f2fe 0%, var(--bg) 45%);
      color: var(--text);
      font: 16px/1.45 "Segoe UI", -apple-system, "Helvetica Neue", sans-serif;
    }
    .wrap {
      max-width: 860px;
      margin: 0 auto;
      padding: 18px 14px 40px;
    }
    .head {
      margin-bottom: 14px;
      color: var(--muted);
      font-size: 14px;
    }
    .toolbar {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
      margin-bottom: 12px;
    }
    .chip {
      display: inline-block;
      padding: 8px 12px;
      border: 1px solid var(--line);
      border-radius: 999px;
      text-decoration: none;
      color: var(--text);
      background: #fff;
      font-size: 14px;
    }
    .chip.active {
      background: var(--accent);
      color: #fff;
      border-color: var(--accent);
    }
    .post {
      background: var(--card);
      border: 1px solid var(--line);
      border-radius: 14px;
      padding: 14px;
      margin-bottom: 12px;
      box-shadow: 0 6px 22px rgba(15, 23, 42, 0.05);
    }
    .meta {
      color: var(--muted);
      font-size: 13px;
      margin-bottom: 8px;
    }
    .txt {
      white-space: pre-wrap;
      margin: 0;
    }
    .empty {
      color: var(--muted);
      font-style: italic;
    }
    .stats {
      margin-top: 10px;
      color: var(--muted);
      font-size: 13px;
    }
    .links {
      margin-top: 10px;
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
    }
    .links a {
      color: var(--accent);
      text-decoration: none;
      font-weight: 600;
    }
    .comments {
      margin-top: 12px;
      border-top: 1px dashed var(--line);
      padding-top: 10px;
      display: grid;
      gap: 8px;
    }
    .comment {
      border: 1px solid var(--line);
      border-radius: 10px;
      padding: 8px 10px;
      background: #fcfdff;
    }
    .comment-meta {
      color: var(--muted);
      font-size: 12px;
      margin-bottom: 4px;
    }
    .comment-text {
      margin: 0;
      white-space: pre-wrap;
      font-size: 14px;
    }
    .media {
      margin-top: 12px;
      display: grid;
      gap: 10px;
    }
    .media-item {
      border: 1px solid var(--line);
      border-radius: 10px;
      padding: 10px;
      background: #f8fafc;
      font-size: 14px;
    }
    .media-item img {
      display: block;
      width: 100%;
      max-height: 420px;
      object-fit: cover;
      border-radius: 8px;
      border: 1px solid var(--line);
      margin-bottom: 8px;
    }
    .media-item a {
      color: var(--accent);
      text-decoration: none;
      font-weight: 600;
      word-break: break-all;
    }
    .more {
      display: inline-block;
      background: var(--accent);
      color: #fff;
      text-decoration: none;
      border-radius: 10px;
      padding: 10px 14px;
      font-weight: 600;
    }
  </style>
</head>
<body>
  <main class="wrap">
    <div class="head">Mudro feed, лимит: {{.Limit}}, страница: {{.Page}}, источник: {{.SourceName}}, сортировка: {{if eq .SortOrder "asc"}}старые сверху{{else}}новые сверху{{end}}</div>
    <div class="head">Всего постов: {{.TotalPosts}} | TG: {{.TGTotal}} | VK: {{.VKTotal}}</div>
    <div class="toolbar">
      <a class="chip {{if eq .Source ""}}active{{end}}" href="{{.AllURL}}">Общая</a>
      <a class="chip {{if eq .Source "vk"}}active{{end}}" href="{{.VKURL}}">VK</a>
      <a class="chip {{if eq .Source "tg"}}active{{end}}" href="{{.TGURL}}">TG</a>
    </div>
    <div class="toolbar">
      <a class="chip {{if eq .SortOrder "desc"}}active{{end}}" href="{{.NewestURL}}">Новые</a>
      <a class="chip {{if eq .SortOrder "asc"}}active{{end}}" href="{{.OldestURL}}">Старые</a>
    </div>
    {{range .Items}}
      <article class="post">
        <div class="meta">#{{.ID}} | {{.Source}}/{{.SourcePostID}} | {{.PublishedAt}}</div>
        {{if .Text}}
          <p class="txt">{{.Text}}</p>
        {{else}}
          <p class="empty">Без текста</p>
        {{end}}
        <div class="stats">likes: {{.LikesCount}} | views: {{if .ViewsCount}}{{.ViewsCount}}{{else}}-{{end}} | comments: {{if .CommentsCount}}{{.CommentsCount}}{{else}}-{{end}}</div>
        {{if .Reactions}}
          <div class="links">
            {{range .Reactions}}
              <span class="chip" title="{{.Raw}}">{{.Label}} {{.Count}}</span>
            {{end}}
          </div>
        {{end}}
        <div class="links">
          {{if .OriginalURL}}<a href="{{.OriginalURL}}" target="_blank" rel="noopener noreferrer">Оригинальный пост</a>{{end}}
          {{if .CommentsURL}}<a href="{{.CommentsURL}}" target="_blank" rel="noopener noreferrer">Обсуждение</a>{{end}}
        </div>
        {{if .Comments}}
          <div class="comments">
            <div class="meta">Комментарии: {{len .Comments}}</div>
            {{range .Comments}}
              <div class="comment">
                <div class="comment-meta">
                  {{if .AuthorName}}{{.AuthorName}}{{else}}без имени{{end}} | {{.PublishedAt}}
                  {{if .ParentCommentID}} | ответ на #{{.ParentCommentID}}{{end}}
                </div>
                {{if .Text}}
                  <p class="comment-text">{{.Text}}</p>
                {{else}}
                  <p class="empty">Без текста</p>
                {{end}}
                {{if .Reactions}}
                  <div class="links">
                    {{range .Reactions}}
                      <span class="chip" title="{{.Raw}}">{{.Label}} {{.Count}}</span>
                    {{end}}
                  </div>
                {{end}}
                {{if .Media}}
                  <div class="media">
                    {{range .Media}}
                      <div class="media-item">
                        {{if .IsImage}}
                          <img src="{{.URL}}" alt="comment media {{.Kind}}" loading="lazy">
                          <a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть изображение</a>
                        {{else if .IsVideo}}
                          {{if .PreviewURL}}<img src="{{.PreviewURL}}" alt="comment video preview" loading="lazy">{{end}}
                          {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть видео</a>{{else}}Видео без URL{{end}}
                        {{else if .IsDocument}}
                          {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть вложение</a>{{else}}Документ без URL{{end}}
                        {{else if .IsLink}}
                          <a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть ссылку</a>
                        {{else}}
                          {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть вложение</a>{{else}}Вложение {{.Kind}}{{end}}
                        {{end}}
                      </div>
                    {{end}}
                  </div>
                {{end}}
              </div>
            {{end}}
          </div>
        {{end}}
        {{if .Media}}
          <div class="media">
            {{range .Media}}
              <div class="media-item">
                {{if .IsImage}}
                  <img src="{{.URL}}" alt="media {{.Kind}}" loading="lazy">
                  <a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть изображение</a>
                {{else if .IsVideo}}
                  {{if .PreviewURL}}<img src="{{.PreviewURL}}" alt="video preview" loading="lazy">{{end}}
                  {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть видео</a>{{else if .Title}}Видео: {{.Title}}{{else}}Видео (без URL в экспорте){{end}}
                {{else if .IsAudio}}
                  {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть аудио</a>{{else if .Title}}Аудио: {{.Title}}{{else}}Аудио во вложении (без URL в экспорте){{end}}
                {{else if .IsDocument}}
                  {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть документ</a>{{else if .Title}}Документ: {{.Title}}{{else}}Документ (без URL){{end}}
                {{else if .IsLink}}
                  {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Связанная ссылка</a>{{else}}Связанная ссылка (без URL){{end}}
                {{else}}
                  {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть вложение ({{.Kind}})</a>{{else if .Title}}Вложение {{.Kind}}: {{.Title}}{{else}}Вложение: {{.Kind}}{{end}}
                {{end}}
              </div>
            {{end}}
          </div>
        {{end}}
      </article>
    {{else}}
      <p class="empty">Постов пока нет.</p>
    {{end}}
    {{if .NextURL}}
      <a class="more" href="{{.NextURL}}">Загрузить еще</a>
    {{end}}
  </main>
</body>
</html>`))

func parseLimit(raw string) int {
	const (
		def = 50
		max = 200
	)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

func parsePage(raw string) (*int, error) {
	if raw == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return nil, errors.New("page must be a positive integer")
	}
	return &n, nil
}

func parseCursor(tsRaw, idRaw string) (*time.Time, *int64, error) {
	if tsRaw == "" && idRaw == "" {
		return nil, nil, nil
	}
	if tsRaw == "" || idRaw == "" {
		return nil, nil, errors.New("both before_ts and before_id are required")
	}
	ts, err := time.Parse(time.RFC3339, tsRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("before_ts: %w", err)
	}
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("before_id: %w", err)
	}
	return &ts, &id, nil
}

func parseSource(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "all":
		return "", nil
	case "vk":
		return "vk", nil
	case "tg":
		return "tg", nil
	default:
		return "", errors.New("use all|vk|tg")
	}
}

func sourceLabel(source string) string {
	if source == "" {
		return "all"
	}
	return source
}

func parseSort(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "desc":
		return "desc", nil
	case "asc":
		return "asc", nil
	default:
		return "", errors.New("use asc|desc")
	}
}

func buildFeedURL(limit, page int, source, sortOrder string) string {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("page", strconv.Itoa(page))
	if source != "" {
		q.Set("source", source)
	}
	if sortOrder != "" {
		q.Set("sort", sortOrder)
	}
	return "/feed?" + q.Encode()
}

func buildOriginalPostURL(source, sourcePostID string) string {
	switch source {
	case "vk":
		if strings.Contains(sourcePostID, "_") {
			return "https://vk.com/wall" + sourcePostID
		}
	}
	return ""
}

func parseMediaItems(raw json.RawMessage) []feedMediaItem {
	items := mediadb.ParseLegacyJSON(raw)
	if len(items) == 0 {
		return nil
	}

	out := make([]feedMediaItem, 0, len(items))
	for _, item := range items {
		kindRaw := strings.ToLower(strings.TrimSpace(item.Kind))
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = guessMediaTitle(item.URL)
		}

		kind := normalizeMediaKind(kindRaw, anyString(item.Extra, "media_type"), anyString(item.Extra, "mime_type"), item.URL, title)
		url := normalizeMediaURL(item.URL)
		preview := normalizeMediaURL(item.PreviewURL)
		if kind == "" && url == "" && preview == "" && title == "" {
			continue
		}

		out = append(out, feedMediaItem{
			Kind:       kind,
			URL:        url,
			PreviewURL: preview,
			Title:      title,
			Width:      item.Width,
			Height:     item.Height,
			Position:   item.Position,
			IsImage:    kind == "photo" || kind == "gif" || kind == "image",
			IsAudio:    kind == "audio",
			IsVideo:    kind == "video",
			IsDocument: kind == "doc",
			IsLink:     kind == "link",
		})
	}
	return out
}

func normalizePostMediaJSON(raw json.RawMessage) json.RawMessage {
	items := parseMediaItems(raw)
	if len(items) == 0 {
		return nil
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return raw
	}
	return json.RawMessage(encoded)
}

func (s *Server) loadPosts(ctx context.Context, beforeTS *time.Time, beforeID *int64, page *int, limit int, source string, sortOrder string) ([]posts.Post, *posts.Cursor, error) {
	if s.postsSvc == nil {
		return nil, nil, nil
	}
	return s.postsSvc.LoadPosts(ctx, beforeTS, beforeID, page, limit, source, posts.SortOrder(sortOrder), "")
}

func (s *Server) loadSourceStats(ctx context.Context) ([]posts.SourceStat, error) {
	if s.postsSvc == nil {
		return nil, nil
	}
	return s.postsSvc.LoadSourceStats(ctx)
}

func (s *Server) countVisiblePosts(ctx context.Context, total *int64) error {
	if s.postsSvc == nil {
		if total != nil {
			*total = 0
		}
		return nil
	}
	count, err := s.postsSvc.CountVisiblePosts(ctx)
	if err != nil {
		return err
	}
	if total != nil {
		*total = count
	}
	return nil
}

func anyString(m map[string]any, keys ...string) string {
	if len(m) == 0 {
		return ""
	}
	for _, key := range keys {
		raw, ok := m[key]
		if !ok || raw == nil {
			continue
		}
		switch v := raw.(type) {
		case string:
			return strings.TrimSpace(v)
		case []byte:
			return strings.TrimSpace(string(v))
		default:
			return strings.TrimSpace(fmt.Sprint(v))
		}
	}
	return ""
}

func normalizeMediaKind(kindRaw, mediaType, mimeType, url, title string) string {
	candidates := []string{kindRaw, mediaType, mimeType, url, title}
	for _, candidate := range candidates {
		kind := strings.ToLower(strings.TrimSpace(candidate))
		if kind == "" {
			continue
		}
		switch {
		case strings.Contains(kind, "video"):
			return "video"
		case strings.Contains(kind, "audio"):
			return "audio"
		case strings.Contains(kind, "gif") || strings.Contains(kind, "sticker"):
			return "gif"
		case strings.Contains(kind, "photo") || strings.Contains(kind, "image"):
			return "photo"
		case strings.Contains(kind, "doc") || strings.Contains(kind, "document") || strings.HasSuffix(kind, ".pdf"):
			return "doc"
		case strings.Contains(kind, "link") || strings.Contains(kind, "url"):
			return "link"
		case strings.Contains(kind, "mp4") || strings.Contains(kind, "webm"):
			return "video"
		case strings.Contains(kind, "mp3") || strings.Contains(kind, "m4a") || strings.Contains(kind, "ogg"):
			return "audio"
		case strings.Contains(kind, "jpg") || strings.Contains(kind, "jpeg") || strings.Contains(kind, "png") || strings.Contains(kind, "webp"):
			return "photo"
		}
	}
	lowerURL := strings.ToLower(strings.TrimSpace(url))
	switch {
	case strings.Contains(lowerURL, ".mp4"), strings.Contains(lowerURL, ".webm"):
		return "video"
	case strings.Contains(lowerURL, ".mp3"), strings.Contains(lowerURL, ".m4a"), strings.Contains(lowerURL, ".ogg"):
		return "audio"
	case strings.Contains(lowerURL, ".pdf"):
		return "doc"
	case strings.Contains(lowerURL, ".jpg"), strings.Contains(lowerURL, ".jpeg"), strings.Contains(lowerURL, ".png"), strings.Contains(lowerURL, ".webp"), strings.Contains(lowerURL, ".gif"):
		return "photo"
	}
	return ""
}

func guessMediaTitle(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if idx := strings.IndexAny(s, "?#"); idx >= 0 {
		s = s[:idx]
	}
	s = strings.TrimRight(s, "/")
	if s == "" {
		return ""
	}
	base := filepath.Base(s)
	if base == "." || base == "/" || base == "" {
		return ""
	}
	return base
}

func normalizeMediaURL(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	if strings.Contains(s, "://") {
		return ""
	}
	if strings.HasPrefix(s, "/") {
		return s
	}
	return "/media/" + s
}

func (s *Server) buildPostsVisibilityWhere(source string, query *string) (string, []any) {
	conditions := make([]string, 0, 3)
	args := []any{}
	if source != "" {
		args = append(args, source)
		conditions = append(conditions, fmt.Sprintf("source = $%d", len(args)))
	}
	if query != nil {
		q := strings.TrimSpace(*query)
		if q != "" {
			args = append(args, "%"+strings.ToLower(q)+"%")
			conditions = append(conditions, fmt.Sprintf("LOWER(text) LIKE $%d", len(args)))
		}
	}
	if len(s.tgVisiblePostIDs) > 0 && (source == "" || source == "tg") {
		args = append(args, s.tgVisiblePostIDs)
		switch source {
		case "tg":
			conditions = append(conditions, fmt.Sprintf("source_post_id = any($%d)", len(args)))
		case "":
			conditions = append(conditions, fmt.Sprintf("(source <> 'tg' or source_post_id = any($%d))", len(args)))
		}
	}
	if len(conditions) == 0 {
		return "", args
	}
	return " where " + strings.Join(conditions, " and "), args
}
