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
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	commentdb "github.com/goritskimihail/mudro/internal/commentmodel"
	"github.com/goritskimihail/mudro/internal/config"
	mediadb "github.com/goritskimihail/mudro/internal/media"
	"github.com/goritskimihail/mudro/internal/tgexport"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	pool             *pgxpool.Pool
	tgVisiblePostIDs []string
}

func NewServer(pool *pgxpool.Pool) *Server {
	server := &Server{pool: pool}
	if ids, path, err := tgexport.LoadVisibleSourcePostIDsFromRepo(config.RepoRoot()); err == nil && len(ids) > 0 {
		server.tgVisiblePostIDs = ids
		log.Printf("api: loaded telegram visibility filter (%d ids) from %s", len(ids), path)
	} else if err != nil {
		log.Printf("api: telegram visibility filter disabled: %v", err)
	}
	return server
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
	mux.HandleFunc("/feed", s.handleFeed)
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(r.Header.Get("Origin")) != "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
	if err := s.countVisiblePosts(ctx, &totalPosts); err != nil {
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
		Items:      make([]feedItem, 0, len(posts)),
		SourceName: sourceLabel(source),
		TotalPosts: totalPosts,
		VKTotal:    vkTotal,
		TGTotal:    tgTotal,
	}

	postIDs := make([]int64, 0, len(posts))
	for _, p := range posts {
		postIDs = append(postIDs, p.ID)
	}
	commentsByPost, err := s.loadPostComments(ctx, postIDs)
	if err != nil {
		log.Printf("feed load comments: %v", err)
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
			Reactions:     buildFeedReactions(p.Reactions),
			Comments:      commentsByPost[p.ID],
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
	args := []any{}
	whereSQL, args := s.buildPostsVisibilityWhere("", args)
	rows, err := s.pool.Query(ctx, `
		select source, count(*) as posts
		from posts`+whereSQL+`
		group by source
		order by posts desc, source asc
	`, args...)
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

func (s *Server) countVisiblePosts(ctx context.Context, out *int64) error {
	args := []any{}
	whereSQL, args := s.buildPostsVisibilityWhere("", args)
	return s.pool.QueryRow(ctx, `select count(*) from posts`+whereSQL, args...).Scan(out)
}

func (s *Server) buildPostsVisibilityWhere(source string, args []any) (string, []any) {
	conditions := make([]string, 0, 2)
	if source != "" {
		args = append(args, source)
		conditions = append(conditions, fmt.Sprintf("source = $%d", len(args)))
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

func (s *Server) loadPosts(ctx context.Context, beforeTS *time.Time, beforeID *int64, page *int, limit int, source, sortOrder string) ([]postDTO, *cursor, error) {
	order := "desc"
	if sortOrder == "asc" {
		order = "asc"
	}
	args := []any{}
	q := `
		select id, source, source_post_id, published_at, text, media, likes_count, views_count, comments_count,
		       coalesce((select count(*)::int from post_comments pc where pc.post_id = posts.id), 0) as actual_comments,
		       created_at, updated_at
		from posts
	`
	whereSQL, nextArgs := s.buildPostsVisibilityWhere(source, args)
	args = nextArgs
	q += whereSQL
	q += fmt.Sprintf(" order by published_at %s, id %s", order, order)
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	posts := make([]postDTO, 0, limit)

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
			&p.ActualComments,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		if len(mediaBytes) > 0 {
			p.Media = json.RawMessage(mediaBytes)
		}
		posts = append(posts, p)
	}
	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}

	posts = dedupeFeedPosts(posts)
	posts, hasMore := slicePosts(posts, beforeTS, beforeID, page, limit, sortOrder)
	if len(posts) == 0 {
		return posts, nil, nil
	}

	ids := make([]int64, 0, len(posts))
	for _, p := range posts {
		ids = append(ids, p.ID)
	}

	normalizedPostMedia, err := mediadb.LoadPostMediaJSON(ctx, s.pool, ids)
	if err != nil {
		return nil, nil, err
	}
	for i := range posts {
		if raw, ok := normalizedPostMedia[posts[i].ID]; ok && len(raw) > 0 {
			posts[i].Media = raw
			continue
		}
		if len(posts[i].Media) > 0 {
			posts[i].Media = normalizePostMediaJSON(posts[i].Media)
		}
	}

	if err := s.loadReactions(ctx, posts, ids); err != nil {
		return nil, nil, err
	}
	commentsByPost, err := s.loadPostComments(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	for i := range posts {
		comments := commentsByPost[posts[i].ID]
		posts[i].Comments = comments
		if posts[i].CommentsCount == nil && len(comments) > 0 {
			count := len(comments)
			posts[i].CommentsCount = &count
		}
	}

	var next *cursor
	if page == nil && hasMore {
		last := posts[len(posts)-1]
		next = &cursor{
			BeforeTS: last.PublishedAt,
			BeforeID: last.ID,
		}
	}
	return posts, next, nil
}

func slicePosts(posts []postDTO, beforeTS *time.Time, beforeID *int64, page *int, limit int, sortOrder string) ([]postDTO, bool) {
	filtered := posts
	if page == nil && beforeTS != nil && beforeID != nil {
		filtered = make([]postDTO, 0, len(posts))
		for _, post := range posts {
			if sortOrder == "asc" {
				if post.PublishedAt.After(*beforeTS) || (post.PublishedAt.Equal(*beforeTS) && post.ID > *beforeID) {
					filtered = append(filtered, post)
				}
				continue
			}
			if post.PublishedAt.Before(*beforeTS) || (post.PublishedAt.Equal(*beforeTS) && post.ID < *beforeID) {
				filtered = append(filtered, post)
			}
		}
	}

	start := 0
	if page != nil {
		start = (*page - 1) * limit
		if start >= len(filtered) {
			return nil, false
		}
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], end < len(filtered)
}

func dedupeFeedPosts(posts []postDTO) []postDTO {
	type recentEntry struct {
		index       int
		publishedAt time.Time
	}

	kept := make([]postDTO, 0, len(posts))
	recentByText := make(map[string][]recentEntry)
	for _, post := range posts {
		key := feedDedupeKey(post)
		if key == "" {
			kept = append(kept, post)
			continue
		}

		recent := recentByText[key]
		pruned := recent[:0]
		duplicateIdx := -1
		for _, entry := range recent {
			if absDurationSeconds(post.PublishedAt.Sub(entry.publishedAt)) > 5 {
				continue
			}
			pruned = append(pruned, entry)
			if duplicateIdx == -1 {
				duplicateIdx = entry.index
			}
		}
		recentByText[key] = pruned
		if duplicateIdx == -1 {
			kept = append(kept, post)
			recentByText[key] = append(recentByText[key], recentEntry{
				index:       len(kept) - 1,
				publishedAt: post.PublishedAt,
			})
			continue
		}
		if feedPostScore(post) > feedPostScore(kept[duplicateIdx]) {
			kept[duplicateIdx] = post
			for i := range recentByText[key] {
				if recentByText[key][i].index == duplicateIdx {
					recentByText[key][i].publishedAt = post.PublishedAt
				}
			}
		}
	}
	return kept
}

func feedDedupeKey(post postDTO) string {
	if post.Source != "tg" || post.Text == nil {
		return ""
	}
	text := normalizeLookupText(*post.Text)
	if text == "" {
		return ""
	}
	return text
}

func feedPostScore(post postDTO) int {
	score := post.ActualComments * 100000
	score += intValue(post.CommentsCount) * 1000
	if len(post.Media) > 0 {
		score += 100
	}
	score += post.LikesCount
	score += len(strings.TrimSpace(compactText(post.Text)))
	return score
}

func absDurationSeconds(d time.Duration) int64 {
	if d < 0 {
		d = -d
	}
	return int64(d / time.Second)
}

func normalizeLookupText(raw string) string {
	text := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(raw, "\r\n", "\n")))
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return text
}

func intValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
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
	ID             int64           `json:"id"`
	Source         string          `json:"source"`
	SourcePostID   string          `json:"source_post_id"`
	PublishedAt    time.Time       `json:"published_at"`
	Text           *string         `json:"text"`
	Media          json.RawMessage `json:"media,omitempty"`
	LikesCount     int             `json:"likes_count"`
	ViewsCount     *int            `json:"views_count,omitempty"`
	CommentsCount  *int            `json:"comments_count,omitempty"`
	Comments       []feedComment   `json:"comments,omitempty"`
	Reactions      map[string]int  `json:"reactions"`
	ActualComments int             `json:"-"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
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

func (s *Server) loadPostComments(ctx context.Context, postIDs []int64) (map[int64][]feedComment, error) {
	out := make(map[int64][]feedComment, len(postIDs))
	if len(postIDs) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `
		select id, post_id, source_comment_id, source_parent_comment_id, parent_comment_id, author_name, published_at, text, reactions, media
		from post_comments
		where post_id = any($1)
		order by post_id asc, published_at asc, id asc
	`, postIDs)
	if err != nil {
		if mediadb.IsUndefinedTableErr(err) {
			return out, nil
		}
		return nil, err
	}
	defer rows.Close()

	type commentRow struct {
		commentID             int64
		postID                int64
		sourceCommentID       string
		sourceParentCommentID *string
		parentCommentRowID    *int64
		authorName            *string
		publishedAt           time.Time
		text                  *string
		reactions             map[string]int
		mediaRaw              json.RawMessage
	}

	staged := make([]commentRow, 0, len(postIDs))
	commentIDs := make([]int64, 0, len(postIDs))
	for rows.Next() {
		var (
			row              commentRow
			reactionsRawJSON []byte
			mediaRawJSON     []byte
		)
		if err := rows.Scan(&row.commentID, &row.postID, &row.sourceCommentID, &row.sourceParentCommentID, &row.parentCommentRowID, &row.authorName, &row.publishedAt, &row.text, &reactionsRawJSON, &mediaRawJSON); err != nil {
			return nil, err
		}
		row.reactions = parseReactionsJSON(reactionsRawJSON)
		if len(mediaRawJSON) > 0 {
			row.mediaRaw = json.RawMessage(mediaRawJSON)
		}
		staged = append(staged, row)
		commentIDs = append(commentIDs, row.commentID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	normalizedCommentMedia, err := mediadb.LoadCommentMediaJSON(ctx, s.pool, commentIDs)
	if err != nil {
		return nil, err
	}
	normalizedCommentReactions, err := commentdb.LoadCommentReactions(ctx, s.pool, commentIDs)
	if err != nil {
		return nil, err
	}

	commentSourceIDs := make(map[int64]string, len(staged))
	for _, row := range staged {
		commentSourceIDs[row.commentID] = row.sourceCommentID
	}

	for _, row := range staged {
		mediaRaw := row.mediaRaw
		if normalized, ok := normalizedCommentMedia[row.commentID]; ok && len(normalized) > 0 {
			mediaRaw = normalized
		}
		reactions := row.reactions
		if normalized, ok := normalizedCommentReactions[row.commentID]; ok && len(normalized) > 0 {
			reactions = normalized
		}
		item := feedComment{
			SourceCommentID: row.sourceCommentID,
			AuthorName:      compactText(row.authorName),
			PublishedAt:     row.publishedAt.Format("2006-01-02 15:04:05"),
			Text:            compactText(row.text),
			Reactions:       buildFeedReactions(reactions),
			Media:           parseMediaItems(mediaRaw),
		}
		switch {
		case row.parentCommentRowID != nil:
			if sourceID := commentSourceIDs[*row.parentCommentRowID]; sourceID != "" {
				item.ParentCommentID = sourceID
			} else if row.sourceParentCommentID != nil {
				item.ParentCommentID = *row.sourceParentCommentID
			}
		case row.sourceParentCommentID != nil:
			item.ParentCommentID = *row.sourceParentCommentID
		}
		out[row.postID] = append(out[row.postID], item)
	}
	return out, nil
}

func parseReactionsJSON(raw []byte) map[string]int {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]int
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func sourceTotals(stats []sourceStat) (vkTotal, tgTotal int64) {
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

func mediaExtraString(m map[string]any, keys ...string) string {
	raw, ok := m["extra"]
	if !ok {
		raw, ok = m["Extra"]
	}
	if !ok {
		return ""
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return ""
	}
	return anyString(obj, keys...)
}

func normalizeMediaURL(raw string) string {
	s := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if s == "" || strings.HasPrefix(s, "missing://") {
		return ""
	}
	if strings.HasPrefix(s, "/media/") {
		return s
	}
	u, err := url.Parse(s)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		return s
	}
	if strings.Contains(s, "://") {
		return ""
	}
	if strings.HasPrefix(s, "/") {
		return s
	}
	return "/media/" + strings.TrimPrefix(strings.TrimPrefix(s, "./"), "/")
}

func guessMediaTitle(rawURL string) string {
	s := strings.TrimSpace(rawURL)
	if s == "" || strings.HasPrefix(s, "missing://") {
		return ""
	}
	if u, err := url.Parse(s); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		return ""
	}
	name := filepath.Base(s)
	if name == "." || name == "/" || strings.TrimSpace(name) == "" {
		return ""
	}
	return name
}

func normalizeMediaKind(kindRaw, mediaTypeRaw, mimeRaw, rawURL, title string) string {
	kind := strings.ToLower(strings.TrimSpace(kindRaw))
	switch kind {
	case "photo", "image", "gif", "video", "audio", "doc", "link":
		return kind
	case "document":
		return "doc"
	}

	mediaType := strings.ToLower(strings.TrimSpace(mediaTypeRaw))
	if strings.Contains(mediaType, "video") {
		return "video"
	}
	if strings.Contains(mediaType, "audio") || strings.Contains(mediaType, "voice") {
		return "audio"
	}
	if strings.Contains(mediaType, "sticker") || strings.Contains(mediaType, "animation") || strings.Contains(mediaType, "gif") {
		return "gif"
	}

	mimeType := strings.ToLower(strings.TrimSpace(mimeRaw))
	if strings.HasPrefix(mimeType, "image/") {
		return "photo"
	}
	if strings.HasPrefix(mimeType, "video/") {
		return "video"
	}
	if strings.HasPrefix(mimeType, "audio/") {
		return "audio"
	}
	if strings.HasPrefix(mimeType, "application/") || strings.HasPrefix(mimeType, "text/") {
		return "doc"
	}

	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(rawURL)))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".bmp", ".tiff":
		return "photo"
	case ".gif":
		return "gif"
	case ".mp4", ".mov", ".mkv", ".webm", ".avi":
		return "video"
	case ".mp3", ".m4a", ".aac", ".ogg", ".wav", ".flac":
		return "audio"
	case ".pdf", ".doc", ".docx", ".txt", ".zip", ".rar", ".7z":
		return "doc"
	}

	if strings.HasPrefix(strings.TrimSpace(rawURL), "http://") || strings.HasPrefix(strings.TrimSpace(rawURL), "https://") {
		if kind == "link" {
			return "link"
		}
		return "doc"
	}
	if strings.TrimSpace(title) != "" && kind == "" {
		return "doc"
	}
	return kind
}
