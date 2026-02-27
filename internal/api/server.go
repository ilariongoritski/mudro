package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	pool *pgxpool.Pool
}

func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{pool: pool}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/posts", s.handlePosts)
	mux.HandleFunc("/api/front", s.handleFront)
	mux.HandleFunc("/feed", s.handleFeed)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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

	posts, next, err := s.loadPosts(ctx, cursorTS, cursorID, page, limit, source, sortOrder)
	if err != nil {
		log.Printf("loadPosts: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := postsResponse{
		Page:       page,
		Limit:      limit,
		Items:      posts,
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

	posts, next, err := s.loadPosts(ctx, nil, nil, nil, limit, source, sortOrder)
	if err != nil {
		log.Printf("loadPosts(front): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var (
		totalPosts int64
		lastSync   *time.Time
	)
	if err := s.pool.QueryRow(ctx, `select count(*) from posts`).Scan(&totalPosts); err != nil {
		log.Printf("front count posts: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := s.pool.QueryRow(ctx, `select max(updated_at) from posts`).Scan(&lastSync); err != nil {
		log.Printf("front max(updated_at): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sources, err := s.loadSourceStats(ctx)
	if err != nil {
		log.Printf("front source stats: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := frontResponse{
		Meta: frontMeta{
			TotalPosts: totalPosts,
			LastSyncAt: lastSync,
			Sources:    sources,
		},
		Feed: postsResponse{
			Limit:      limit,
			Items:      posts,
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

	posts, _, err := s.loadPosts(ctx, nil, nil, page, limit, source, sortOrder)
	if err != nil {
		log.Printf("loadPosts(feed): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

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
		Items:      make([]feedItem, 0, len(posts)),
		SourceName: sourceLabel(source),
	}
	for _, p := range posts {
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
			Media:         parseMediaItems(p.Media),
		})
	}

	if len(posts) == limit {
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

func (s *Server) loadSourceStats(ctx context.Context) ([]sourceStat, error) {
	rows, err := s.pool.Query(ctx, `
		select source, count(*) as posts
		from posts
		group by source
		order by posts desc, source asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []sourceStat
	for rows.Next() {
		var st sourceStat
		if err := rows.Scan(&st.Source, &st.Posts); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *Server) loadPosts(ctx context.Context, beforeTS *time.Time, beforeID *int64, page *int, limit int, source, sortOrder string) ([]postDTO, *cursor, error) {
	var (
		rows pgx.Rows
		err  error
	)
	order := "desc"
	comparator := "<"
	if sortOrder == "asc" {
		order = "asc"
		comparator = ">"
	}

	if page != nil {
		offset := (*page - 1) * limit
		args := []any{}
		q := `
			select id, source, source_post_id, published_at, text, media, likes_count, views_count, comments_count, created_at, updated_at
			from posts
		`
		if source != "" {
			args = append(args, source)
			q += fmt.Sprintf(" where source = $%d", len(args))
		}
		args = append(args, limit, offset)
		q += fmt.Sprintf(" order by published_at %s, id %s limit $%d offset $%d", order, order, len(args)-1, len(args))
		rows, err = s.pool.Query(ctx, q, args...)
	} else {
		base := `
			select id, source, source_post_id, published_at, text, media, likes_count, views_count, comments_count, created_at, updated_at
			from posts
		`
		if beforeTS == nil || beforeID == nil {
			args := []any{}
			q := base
			if source != "" {
				args = append(args, source)
				q += fmt.Sprintf(" where source = $%d", len(args))
			}
			args = append(args, limit)
			q += fmt.Sprintf(" order by published_at %s, id %s limit $%d", order, order, len(args))
			rows, err = s.pool.Query(ctx, q, args...)
		} else {
			args := []any{}
			q := base
			if source != "" {
				args = append(args, source)
				q += fmt.Sprintf(" where source = $%d and ", len(args))
			} else {
				q += " where "
			}
			args = append(args, *beforeTS, *beforeID, limit)
			q += fmt.Sprintf("(published_at, id) %s ($%d, $%d) order by published_at %s, id %s limit $%d", comparator, len(args)-2, len(args)-1, order, order, len(args))
			rows, err = s.pool.Query(ctx, q, args...)
		}
	}
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	posts := make([]postDTO, 0, limit)
	ids := make([]int64, 0, limit)

	for rows.Next() {
		var (
			p          postDTO
			mediaBytes []byte
		)
		if err := rows.Scan(
			&p.ID,
			&p.Source,
			&p.SourcePostID,
			&p.PublishedAt,
			&p.Text,
			&mediaBytes,
			&p.LikesCount,
			&p.ViewsCount,
			&p.CommentsCount,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		if len(mediaBytes) > 0 {
			p.Media = json.RawMessage(mediaBytes)
		}
		posts = append(posts, p)
		ids = append(ids, p.ID)
	}
	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}

	if len(posts) == 0 {
		return posts, nil, nil
	}

	if err := s.loadReactions(ctx, posts, ids); err != nil {
		return nil, nil, err
	}

	var next *cursor
	if page == nil {
		last := posts[len(posts)-1]
		next = &cursor{
			BeforeTS: last.PublishedAt,
			BeforeID: last.ID,
		}
	}
	return posts, next, nil
}

func (s *Server) loadReactions(ctx context.Context, posts []postDTO, ids []int64) error {
	rows, err := s.pool.Query(ctx, `
		select post_id, emoji, count
		from post_reactions
		where post_id = any($1)
	`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	index := make(map[int64]*postDTO, len(posts))
	for i := range posts {
		posts[i].Reactions = map[string]int{}
		index[posts[i].ID] = &posts[i]
	}

	for rows.Next() {
		var (
			postID int64
			emoji  string
			count  int
		)
		if err := rows.Scan(&postID, &emoji, &count); err != nil {
			return err
		}
		if p := index[postID]; p != nil {
			p.Reactions[emoji] = count
		}
	}
	return rows.Err()
}

type postDTO struct {
	ID            int64           `json:"id"`
	Source        string          `json:"source"`
	SourcePostID  string          `json:"source_post_id"`
	PublishedAt   time.Time       `json:"published_at"`
	Text          *string         `json:"text"`
	Media         json.RawMessage `json:"media,omitempty"`
	LikesCount    int             `json:"likes_count"`
	ViewsCount    *int            `json:"views_count,omitempty"`
	CommentsCount *int            `json:"comments_count,omitempty"`
	Reactions     map[string]int  `json:"reactions"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type cursor struct {
	BeforeTS time.Time `json:"before_ts"`
	BeforeID int64     `json:"before_id"`
}

type postsResponse struct {
	Page       *int      `json:"page,omitempty"`
	Limit      int       `json:"limit"`
	Items      []postDTO `json:"items"`
	NextCursor *cursor   `json:"next_cursor,omitempty"`
}

type frontResponse struct {
	Meta frontMeta     `json:"meta"`
	Feed postsResponse `json:"feed"`
}

type frontMeta struct {
	TotalPosts int64        `json:"total_posts"`
	LastSyncAt *time.Time   `json:"last_sync_at,omitempty"`
	Sources    []sourceStat `json:"sources"`
}

type sourceStat struct {
	Source string `json:"source"`
	Posts  int64  `json:"posts"`
}

type feedPageData struct {
	Limit      int
	Page       int
	Source     string
	SourceName string
	SortOrder  string
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
	Media         []feedMediaItem
}

type feedMediaItem struct {
	Kind       string
	URL        string
	PreviewURL string
	Width      int
	Height     int
	Position   int
	IsImage    bool
	IsAudio    bool
	IsVideo    bool
	IsDocument bool
	IsLink     bool
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
        <div class="links">
          {{if .OriginalURL}}<a href="{{.OriginalURL}}" target="_blank" rel="noopener noreferrer">Оригинальный пост</a>{{end}}
          {{if .CommentsURL}}<a href="{{.CommentsURL}}" target="_blank" rel="noopener noreferrer">Обсуждение</a>{{end}}
        </div>
        {{if .Media}}
          <div class="media">
            {{range .Media}}
              <div class="media-item">
                {{if .IsImage}}
                  <img src="{{.URL}}" alt="media {{.Kind}}" loading="lazy">
                  <a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть изображение</a>
                {{else if .IsVideo}}
                  {{if .PreviewURL}}<img src="{{.PreviewURL}}" alt="video preview" loading="lazy">{{end}}
                  {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть видео</a>{{else}}Видео (без URL в экспорте){{end}}
                {{else if .IsAudio}}
                  {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть аудио</a>{{else}}Аудио во вложении (без URL в экспорте){{end}}
                {{else if .IsDocument}}
                  {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть документ</a>{{else}}Документ (без URL){{end}}
                {{else if .IsLink}}
                  {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Связанная ссылка</a>{{else}}Связанная ссылка (без URL){{end}}
                {{else}}
                  {{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть вложение ({{.Kind}})</a>{{else}}Вложение: {{.Kind}}{{end}}
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
	if len(raw) == 0 {
		return nil
	}

	var decoded []map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil
	}

	items := make([]feedMediaItem, 0, len(decoded))
	for _, m := range decoded {
		kind := strings.ToLower(strings.TrimSpace(anyString(m, "kind", "Kind")))
		url := strings.TrimSpace(anyString(m, "url", "URL"))
		preview := strings.TrimSpace(anyString(m, "preview_url", "PreviewURL"))
		width := anyInt(m, "width", "Width")
		height := anyInt(m, "height", "Height")
		position := anyInt(m, "position", "Position")
		if kind == "" && url == "" && preview == "" {
			continue
		}
		items = append(items, feedMediaItem{
			Kind:       kind,
			URL:        url,
			PreviewURL: preview,
			Width:      width,
			Height:     height,
			Position:   position,
			IsImage:    kind == "photo" || kind == "gif",
			IsAudio:    kind == "audio",
			IsVideo:    kind == "video",
			IsDocument: kind == "doc",
			IsLink:     kind == "link",
		})
	}
	return items
}

func anyString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func anyInt(m map[string]any, keys ...string) int {
	for _, k := range keys {
		v, ok := m[k]
		if !ok {
			continue
		}
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case int32:
			return int(n)
		case int64:
			return int(n)
		}
	}
	return 0
}
