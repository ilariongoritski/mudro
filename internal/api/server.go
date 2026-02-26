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

	posts, next, err := s.loadPosts(ctx, cursorTS, cursorID, page, limit)
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

	posts, next, err := s.loadPosts(ctx, nil, nil, nil, limit)
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
	limit := parseLimit(r.URL.Query().Get("limit"))

	cursorTS, cursorID, err := parseCursor(r.URL.Query().Get("before_ts"), r.URL.Query().Get("before_id"))
	if err != nil {
		http.Error(w, "invalid cursor: "+err.Error(), http.StatusBadRequest)
		return
	}

	posts, next, err := s.loadPosts(ctx, cursorTS, cursorID, nil, limit)
	if err != nil {
		log.Printf("loadPosts(feed): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := feedPageData{
		Limit: limit,
		Items: make([]feedItem, 0, len(posts)),
	}
	for _, p := range posts {
		data.Items = append(data.Items, feedItem{
			ID:            p.ID,
			Source:        p.Source,
			SourcePostID:  p.SourcePostID,
			PublishedAt:   p.PublishedAt.Format("2006-01-02 15:04:05"),
			Text:          compactText(p.Text),
			LikesCount:    p.LikesCount,
			ViewsCount:    p.ViewsCount,
			CommentsCount: p.CommentsCount,
		})
	}

	if next != nil {
		q := url.Values{}
		q.Set("limit", strconv.Itoa(limit))
		q.Set("before_ts", next.BeforeTS.Format(time.RFC3339))
		q.Set("before_id", strconv.FormatInt(next.BeforeID, 10))
		data.NextURL = "/feed?" + q.Encode()
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

func (s *Server) loadPosts(ctx context.Context, beforeTS *time.Time, beforeID *int64, page *int, limit int) ([]postDTO, *cursor, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if page != nil {
		offset := (*page - 1) * limit
		rows, err = s.pool.Query(ctx, `
			select id, source, source_post_id, published_at, text, media, likes_count, views_count, comments_count, created_at, updated_at
			from posts
			order by published_at desc, id desc
			limit $1 offset $2
		`, limit, offset)
		goto scan
	}

	if beforeTS == nil || beforeID == nil {
		rows, err = s.pool.Query(ctx, `
			select id, source, source_post_id, published_at, text, media, likes_count, views_count, comments_count, created_at, updated_at
			from posts
			order by published_at desc, id desc
			limit $1
		`, limit)
	} else {
		rows, err = s.pool.Query(ctx, `
			select id, source, source_post_id, published_at, text, media, likes_count, views_count, comments_count, created_at, updated_at
			from posts
			where (published_at, id) < ($1, $2)
			order by published_at desc, id desc
			limit $3
		`, *beforeTS, *beforeID, limit)
	}
	if err != nil {
		return nil, nil, err
	}
scan:
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
	Limit   int
	Items   []feedItem
	NextURL string
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
    <div class="head">Mudro feed, лимит: {{.Limit}}</div>
    {{range .Items}}
      <article class="post">
        <div class="meta">#{{.ID}} | {{.Source}}/{{.SourcePostID}} | {{.PublishedAt}}</div>
        {{if .Text}}
          <p class="txt">{{.Text}}</p>
        {{else}}
          <p class="empty">Без текста</p>
        {{end}}
        <div class="stats">likes: {{.LikesCount}} | views: {{if .ViewsCount}}{{.ViewsCount}}{{else}}-{{end}} | comments: {{if .CommentsCount}}{{.CommentsCount}}{{else}}-{{end}}</div>
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
