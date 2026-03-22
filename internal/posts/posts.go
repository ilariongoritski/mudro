package posts

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	commentdb "github.com/goritskimihail/mudro/internal/commentmodel"
	mediadb "github.com/goritskimihail/mudro/internal/media"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool             *pgxpool.Pool
	tgVisiblePostIDs []string
}

func NewService(pool *pgxpool.Pool, tgVisiblePostIDs []string) *Service {
	return &Service{
		pool:             pool,
		tgVisiblePostIDs: tgVisiblePostIDs,
	}
}

type SortOrder string

const (
	SortDesc SortOrder = "desc"
	SortAsc  SortOrder = "asc"
)

type Cursor struct {
	BeforeTS time.Time `json:"before_ts"`
	BeforeID int64     `json:"before_id"`
}

type Post struct {
	ID            int64           `json:"id"`
	Source        string          `json:"source"`
	SourcePostID  string          `json:"source_post_id"`
	PublishedAt   time.Time       `json:"published_at"`
	Text          *string         `json:"text,omitempty"`
	Media         json.RawMessage `json:"media,omitempty"`
	LikesCount    int             `json:"likes_count"`
	ViewsCount    *int            `json:"views_count,omitempty"`
	CommentsCount *int            `json:"comments_count,omitempty"`
	Reactions     map[string]int  `json:"reactions,omitempty"`
	Comments      []Comment       `json:"comments,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type Comment struct {
	SourceCommentID string          `json:"source_comment_id"`
	ParentCommentID string          `json:"parent_comment_id,omitempty"`
	AuthorName      string          `json:"author_name,omitempty"`
	PublishedAt     string          `json:"published_at"`
	Text            string          `json:"text,omitempty"`
	Reactions       map[string]int  `json:"reactions,omitempty"`
	Media           json.RawMessage `json:"media,omitempty"`
}

type MediaItem struct {
	Kind       string `json:"kind"`
	URL        string `json:"url,omitempty"`
	PreviewURL string `json:"preview_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Position   int    `json:"position,omitempty"`
}

type SourceStat struct {
	Source string `json:"source"`
	Posts  int64  `json:"posts"`
}

type Reaction struct {
	Label string `json:"label"`
	Count int    `json:"count"`
	Raw   string `json:"raw"`
}

func (s *Service) LoadPosts(ctx context.Context, beforeTS *time.Time, beforeID *int64, page *int, limit int, source string, sortOrder SortOrder, query string) ([]Post, *Cursor, error) {
	var (
		rows pgx.Rows
		err  error
	)

	order := "desc"
	comparator := "<"
	if sortOrder == SortAsc {
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
		whereSQL, nextArgs := s.buildPostsVisibilityWhere(source, query, args)
		args = nextArgs
		q += whereSQL
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
			whereSQL, nextArgs := s.buildPostsVisibilityWhere(source, query, args)
			args = nextArgs
			q += whereSQL
			args = append(args, limit)
			q += fmt.Sprintf(" order by published_at %s, id %s limit $%d", order, order, len(args))
			rows, err = s.pool.Query(ctx, q, args...)
		} else {
			args := []any{}
			q := base
			whereSQL, nextArgs := s.buildPostsVisibilityWhere(source, query, args)
			args = nextArgs
			if whereSQL == "" {
				q += " where "
			} else {
				q += whereSQL + " and "
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

	posts := make([]Post, 0, limit)
	ids := make([]int64, 0, limit)
	for rows.Next() {
		var p Post
		if err := rows.Scan(
			&p.ID, &p.Source, &p.SourcePostID, &p.PublishedAt, &p.Text, &p.Media,
			&p.LikesCount, &p.ViewsCount, &p.CommentsCount, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		posts = append(posts, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if len(posts) == 0 {
		return posts, nil, nil
	}

	normalizedPostMedia, err := mediadb.LoadPostMediaJSON(ctx, s.pool, ids)
	if err != nil {
		return nil, nil, err
	}
	postReactions, err := s.loadPostReactions(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	commentsMap, err := s.loadPostComments(ctx, ids)
	if err != nil {
		return nil, nil, err
	}

	for i := range posts {
		if raw, ok := normalizedPostMedia[posts[i].ID]; ok {
			posts[i].Media = raw
		}
		posts[i].Reactions = postReactions[posts[i].ID]
		posts[i].Comments = commentsMap[posts[i].ID]
	}

	last := posts[len(posts)-1]
	return posts, &Cursor{BeforeTS: last.PublishedAt, BeforeID: last.ID}, nil
}

func (s *Service) buildPostsVisibilityWhere(source, query string, args []any) (string, []any) {
	conditions := make([]string, 0, 3)
	if source != "" {
		args = append(args, source)
		conditions = append(conditions, fmt.Sprintf("source = $%d", len(args)))
	}
	if query != "" {
		args = append(args, "%"+strings.ToLower(query)+"%")
		conditions = append(conditions, fmt.Sprintf("LOWER(text) LIKE $%d", len(args)))
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

func (s *Service) loadPostReactions(ctx context.Context, postIDs []int64) (map[int64]map[string]int, error) {
	out := make(map[int64]map[string]int, len(postIDs))
	if len(postIDs) == 0 {
		return out, nil
	}

	rows, err := s.pool.Query(ctx, `
		select post_id, emoji, count
		from post_reactions
		where post_id = any($1)
		order by post_id asc, emoji asc
	`, postIDs)
	if err != nil {
		if commentdb.IsUndefinedTableErr(err) {
			return out, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			postID int64
			emoji  string
			count  int
		)
		if err := rows.Scan(&postID, &emoji, &count); err != nil {
			return nil, err
		}
		if count <= 0 {
			continue
		}
		if out[postID] == nil {
			out[postID] = map[string]int{}
		}
		out[postID][emoji] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) loadPostComments(ctx context.Context, postIDs []int64) (map[int64][]Comment, error) {
	out := make(map[int64][]Comment, len(postIDs))
	if len(postIDs) == 0 {
		return out, nil
	}

	rows, err := s.pool.Query(ctx, `
		select id, post_id, source_comment_id, source_parent_comment_id, author_name, published_at, text, reactions, media
		from post_comments
		where post_id = any($1)
		order by post_id asc, published_at asc, id asc
	`, postIDs)
	if err != nil {
		if commentdb.IsUndefinedTableErr(err) {
			return out, nil
		}
		return nil, err
	}
	defer rows.Close()

	type commentRow struct {
		commentID       int64
		postID          int64
		sourceCommentID string
		parentCommentID *string
		authorName      *string
		publishedAt     time.Time
		text            *string
		reactionsRaw    []byte
		mediaRaw        json.RawMessage
	}

	staged := make([]commentRow, 0, len(postIDs))
	commentIDs := make([]int64, 0, len(postIDs))
	for rows.Next() {
		var row commentRow
		if err := rows.Scan(&row.commentID, &row.postID, &row.sourceCommentID, &row.parentCommentID, &row.authorName, &row.publishedAt, &row.text, &row.reactionsRaw, &row.mediaRaw); err != nil {
			return nil, err
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

	for _, row := range staged {
		reactions := parseReactionsJSON(row.reactionsRaw)
		if nr, ok := normalizedCommentReactions[row.commentID]; ok && len(nr) > 0 {
			reactions = nr
		}

		mediaRaw := row.mediaRaw
		if nm, ok := normalizedCommentMedia[row.commentID]; ok && len(nm) > 0 {
			mediaRaw = nm
		}

		c := Comment{
			SourceCommentID: row.sourceCommentID,
			ParentCommentID: ptrString(row.parentCommentID),
			AuthorName:      ptrString(row.authorName),
			PublishedAt:     row.publishedAt.Format(time.RFC3339),
			Text:            ptrString(row.text),
			Reactions:       reactions,
			Media:           mediaRaw,
		}
		out[row.postID] = append(out[row.postID], c)
	}
	return out, nil
}

func parseReactionsJSON(raw []byte) map[string]int {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]int
	_ = json.Unmarshal(raw, &out)
	return out
}

func ptrString(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func ParseMediaItems(raw json.RawMessage) []MediaItem {
	items := mediadb.ParseLegacyJSON(raw)
	out := make([]MediaItem, 0, len(items))
	for _, item := range items {
		out = append(out, MediaItem{
			Kind:       normalizeMediaKind(item.Kind, anyString(item.Extra, "media_type"), anyString(item.Extra, "mime_type"), item.URL, item.Title),
			URL:        NormalizeMediaURL(item.URL),
			PreviewURL: NormalizeMediaURL(item.PreviewURL),
			Title:      item.Title,
			Width:      item.Width,
			Height:     item.Height,
			Position:   item.Position,
		})
	}
	return out
}

func NormalizeMediaURL(raw string) string {
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

func (s *Service) LoadSourceStats(ctx context.Context) ([]SourceStat, error) {
	args := []any{}
	whereSQL, args := s.buildPostsVisibilityWhere("", "", args)
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

	var out []SourceStat
	for rows.Next() {
		var st SourceStat
		if err := rows.Scan(&st.Source, &st.Posts); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *Service) CountVisiblePosts(ctx context.Context) (int64, error) {
	var count int64
	args := []any{}
	whereSQL, args := s.buildPostsVisibilityWhere("", "", args)
	err := s.pool.QueryRow(ctx, `select count(*) from posts`+whereSQL, args...).Scan(&count)
	return count, err
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
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(url)), "http") {
		if strings.Contains(strings.ToLower(url), ".mp4") || strings.Contains(strings.ToLower(url), ".webm") {
			return "video"
		}
		if strings.Contains(strings.ToLower(url), ".mp3") || strings.Contains(strings.ToLower(url), ".m4a") || strings.Contains(strings.ToLower(url), ".ogg") {
			return "audio"
		}
		if strings.Contains(strings.ToLower(url), ".pdf") {
			return "doc"
		}
		if strings.Contains(strings.ToLower(url), ".jpg") || strings.Contains(strings.ToLower(url), ".jpeg") || strings.Contains(strings.ToLower(url), ".png") || strings.Contains(strings.ToLower(url), ".webp") || strings.Contains(strings.ToLower(url), ".gif") {
			return "photo"
		}
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
	base := path.Base(s)
	if base == "." || base == "/" || base == "" {
		return ""
	}
	return base
}
