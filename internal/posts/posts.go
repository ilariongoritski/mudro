package posts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	commentdb "github.com/goritskimihail/mudro/internal/commentmodel"
	mediadb "github.com/goritskimihail/mudro/internal/media"
	"github.com/goritskimihail/mudro/pkg/models"
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

type SortOrder = models.SortOrder

const (
	SortDesc = models.SortDesc
	SortAsc  = models.SortAsc
)

type Cursor = models.Cursor
type Post = models.Post
type Comment = models.Comment
type MediaItem = models.MediaItem
type SourceStat = models.SourceStat
type Reaction = models.Reaction


func (s *Service) CountVisiblePosts(ctx context.Context) (int64, error) {
	q := `SELECT count(*) FROM posts`
	args := make([]any, 0, 1)
	if len(s.tgVisiblePostIDs) > 0 {
		args = append(args, s.tgVisiblePostIDs)
		q += fmt.Sprintf(` WHERE (source <> 'tg' OR source_post_id = any($%d))`, len(args))
	}
	var count int64
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}


// LoadLastSyncAt returns the most recent updated_at across all posts.
// Exists so HTTP handlers get this via service layer — not direct pool access.
func (s *Service) LoadLastSyncAt(ctx context.Context) (*time.Time, error) {
	var lastSync *time.Time
	if err := s.pool.QueryRow(ctx, `SELECT MAX(updated_at) FROM posts`).Scan(&lastSync); err != nil {
		return nil, err
	}
	return lastSync, nil
}
func (s *Service) LoadSourceStats(ctx context.Context) ([]SourceStat, error) {
	q := `SELECT source, count(*) FROM posts`
	args := make([]any, 0, 1)
	if len(s.tgVisiblePostIDs) > 0 {
		args = append(args, s.tgVisiblePostIDs)
		q += fmt.Sprintf(` WHERE (source <> 'tg' OR source_post_id = any($%d))`, len(args))
	}
	q += ` GROUP BY source ORDER BY source`

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []SourceStat
	for rows.Next() {
		var st SourceStat
		if err := rows.Scan(&st.Source, &st.Posts); err != nil {
			return nil, err
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

func ParseMediaItems(raw json.RawMessage) []MediaItem {
	items := mediadb.ParseLegacyJSON(raw)
	out := make([]MediaItem, 0, len(items))
	for _, item := range items {
		out = append(out, MediaItem{
			Kind:       item.Kind,
			URL:        item.URL,
			PreviewURL: item.PreviewURL,
			Title:      item.Title,
			Width:      item.Width,
			Height:     item.Height,
			Position:   item.Position,
		})
	}
	return out
}

func (s *Service) LoadPosts(ctx context.Context, beforeTS *time.Time, beforeID *int64, page *int, limit int, source string, sortOrder SortOrder, query string) ([]Post, *Cursor, error) {
	order := "desc"
	comparator := "<"
	if sortOrder == SortAsc {
		order = "asc"
		comparator = ">"
	}

	q, args, useCursor := s.buildPostsQuery(source, query, page, beforeTS, beforeID, limit, order, comparator)
	rows, err := s.pool.Query(ctx, q, args...)
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

	if !useCursor {
		return posts, nil, nil
	}

	last := posts[len(posts)-1]
	return posts, &Cursor{BeforeTS: last.PublishedAt, BeforeID: last.ID}, nil
}

func (s *Service) buildPostsQuery(source, query string, page *int, beforeTS *time.Time, beforeID *int64, limit int, order, comparator string) (string, []any, bool) {
	base := `
		select id, source, source_post_id, published_at, text, media, likes_count, views_count, comments_count, created_at, updated_at
		from posts
	`
	args := make([]any, 0, 4)
	whereSQL, args := s.buildPostsVisibilityWhere(source, query, args)
	q := base + whereSQL

	if page != nil {
		offset := (*page - 1) * limit
		args = append(args, limit, offset)
		q += fmt.Sprintf(" order by published_at %s, id %s limit $%d offset $%d", order, order, len(args)-1, len(args))
		return q, args, false
	}

	if beforeTS != nil && beforeID != nil {
		if whereSQL == "" {
			q += " where "
		} else {
			q += " and "
		}
		args = append(args, *beforeTS, *beforeID)
		q += fmt.Sprintf("(published_at, id) %s ($%d, $%d)", comparator, len(args)-1, len(args))
	}

	args = append(args, limit)
	q += fmt.Sprintf(" order by published_at %s, id %s limit $%d", order, order, len(args))
	return q, args, beforeTS != nil && beforeID != nil
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
		select id, post_id, parent_comment_id, author_name, published_at, text, reactions, media
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

	for rows.Next() {
		var (
			commentID       int64
			postID          int64
			parentCommentID *string
			authorName      *string
			publishedAt     time.Time
			text            *string
			reactionsRaw    []byte
			mediaRaw        []byte
		)
		if err := rows.Scan(&commentID, &postID, &parentCommentID, &authorName, &publishedAt, &text, &reactionsRaw, &mediaRaw); err != nil {
			return nil, err
		}

		comment := Comment{SourceCommentID: fmt.Sprintf("%d", commentID), PublishedAt: publishedAt.UTC().Format(time.RFC3339)}
		if parentCommentID != nil {
			comment.ParentCommentID = strings.TrimSpace(*parentCommentID)
		}
		if authorName != nil {
			comment.AuthorName = strings.TrimSpace(*authorName)
		}
		if text != nil {
			comment.Text = *text
		}
		if len(reactionsRaw) > 0 {
			_ = json.Unmarshal(reactionsRaw, &comment.Reactions)
		}
		if len(mediaRaw) > 0 {
			comment.Media = append(json.RawMessage(nil), mediaRaw...)
		}

		out[postID] = append(out[postID], comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}


// AddComment safely inserts a comment to a post.
func (s *Service) AddComment(ctx context.Context, postID int64, authorName, text string, parentCommentID *int64) (int64, time.Time, error) {
	now := time.Now()
	var commentID int64
	err := s.pool.QueryRow(ctx,
		`insert into post_comments (post_id, source, source_comment_id, author_name, text, published_at, parent_comment_id)
		 values ($1, 'local', 'local-' || nextval('post_comments_id_seq')::text, $2, $3, $4, $5)
		 returning id`,
		postID, authorName, text, now, parentCommentID,
	).Scan(&commentID)
	return commentID, now, err
}

// ToggleLike toggles the like status of a user on a post.
func (s *Service) ToggleLike(ctx context.Context, postID, userID int64) (liked bool, likesCount int, err error) {
	tag, err := s.pool.Exec(ctx,
		`insert into post_user_likes (post_id, user_id) values ($1, $2) on conflict do nothing`,
		postID, userID,
	)
	if err != nil {
		return false, 0, err
	}

	if tag.RowsAffected() == 0 {
		_, err = s.pool.Exec(ctx,
			`delete from post_user_likes where post_id = $1 and user_id = $2`,
			postID, userID,
		)
		if err != nil {
			return false, 0, err
		}
		liked = false
	} else {
		liked = true
	}

	_ = s.pool.QueryRow(ctx, `select likes_count from posts where id = $1`, postID).Scan(&likesCount)
	return liked, likesCount, nil
}