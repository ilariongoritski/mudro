package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
