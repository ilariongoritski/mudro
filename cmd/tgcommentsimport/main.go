package main

import (
	"context"
	"encoding/json"
	"flag"
	"html"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type htmlMessage struct {
	ID          int64
	ReplyToID   int64
	FromName    string
	Text        string
	PublishedAt time.Time
	Reactions   map[string]int
}

type tgPostRef struct {
	PostID      int64
	PublishedAt time.Time
}

var (
	reMsgStart  = regexp.MustCompile(`<div class="message default clearfix(?: joined)?" id="message(\d+)">`)
	reDateTitle = regexp.MustCompile(`class="pull_right date details" title="([^"]+)"`)
	reTextBlock = regexp.MustCompile(`(?s)<div class="text">\s*(.*?)\s*</div>`)
	reFromName  = regexp.MustCompile(`(?s)<div class="from_name">\s*(.*?)\s*</div>`)
	reReplyTo   = regexp.MustCompile(`GoToMessage\((\d+)\)`)
	reReaction  = regexp.MustCompile(`(?s)<span class="reaction">.*?<span class="emoji">\s*(.*?)\s*</span>.*?<span class="count">\s*(\d+)\s*</span>.*?</span>`)
	reTags      = regexp.MustCompile(`(?s)<[^>]*>`)
	reSpaces    = regexp.MustCompile(`[ \t\r\f\v]+`)
)

func main() {
	dir := flag.String("dir", "data/nu", "directory with messages*.html")
	dsn := flag.String("dsn", config.DSN(), "postgres dsn")
	flag.Parse()

	files, err := filepath.Glob(filepath.Join(*dir, "messages*.html"))
	if err != nil {
		log.Fatalf("glob html files: %v", err)
	}
	if len(files) == 0 {
		log.Fatalf("no files matched: %s", filepath.Join(*dir, "messages*.html"))
	}
	sort.Strings(files)

	msgs, ids, err := parseHTMLMessages(files)
	if err != nil {
		log.Fatalf("parse html: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	postByTGMsgID, err := loadTGPostMap(ctx, pool)
	if err != nil {
		log.Fatalf("load posts map: %v", err)
	}

	inserted := 0
	createdPosts := 0
	skippedNoReply := 0
	skippedNoRoot := 0
	for _, id := range ids {
		msg := msgs[id]
		if msg.ReplyToID == 0 {
			skippedNoReply++
			continue
		}
		postID, parentCommentID, ok := resolvePostLink(msgs, postByTGMsgID, msg.ReplyToID, msg.PublishedAt)
		if !ok {
			rootID, rootMsg, found := findRootMessage(msgs, msg.ReplyToID)
			if found {
				if _, exists := postByTGMsgID[rootID]; !exists {
					created, upsertedPostID, publishedAt, err := upsertRootPost(context.Background(), pool, rootMsg)
					if err != nil {
						log.Fatalf("upsert root post msg=%d: %v", rootID, err)
					}
					postByTGMsgID[rootID] = tgPostRef{PostID: upsertedPostID, PublishedAt: publishedAt}
					if created {
						createdPosts++
					}
				}
				postID, parentCommentID, ok = resolvePostLink(msgs, postByTGMsgID, msg.ReplyToID, msg.PublishedAt)
			}
		}
		if !ok {
			skippedNoRoot++
			continue
		}
		if err := upsertComment(context.Background(), pool, postID, msg, parentCommentID); err != nil {
			log.Fatalf("upsert comment msg=%d: %v", msg.ID, err)
		}
		inserted++
	}

	log.Printf("DONE: parsed=%d imported_comments=%d created_posts=%d skipped_no_reply=%d skipped_no_root=%d", len(ids), inserted, createdPosts, skippedNoReply, skippedNoRoot)
}

func parseHTMLMessages(files []string) (map[int64]htmlMessage, []int64, error) {
	out := make(map[int64]htmlMessage, 4096)
	ordered := make([]int64, 0, 4096)

	for _, path := range files {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		s := string(b)
		starts := reMsgStart.FindAllStringSubmatchIndex(s, -1)
		if len(starts) == 0 {
			continue
		}

		lastFromName := ""
		for i, st := range starts {
			blockStart := st[0]
			blockEnd := len(s)
			if i+1 < len(starts) {
				blockEnd = starts[i+1][0]
			}
			block := s[blockStart:blockEnd]

			msgID, err := strconv.ParseInt(s[st[2]:st[3]], 10, 64)
			if err != nil {
				continue
			}
			dateMatch := reDateTitle.FindStringSubmatch(block)
			if len(dateMatch) < 2 {
				continue
			}
			publishedAt, err := time.Parse("02.01.2006 15:04:05 UTC-07:00", strings.TrimSpace(dateMatch[1]))
			if err != nil {
				continue
			}

			fromName := ""
			if m := reFromName.FindStringSubmatch(block); len(m) >= 2 {
				fromName = cleanHTMLText(m[1])
				if fromName != "" {
					lastFromName = fromName
				}
			}
			if fromName == "" {
				fromName = lastFromName
			}

			text := ""
			if m := reTextBlock.FindStringSubmatch(block); len(m) >= 2 {
				text = cleanHTMLText(m[1])
			}

			replyTo := int64(0)
			if m := reReplyTo.FindStringSubmatch(block); len(m) >= 2 {
				if v, err := strconv.ParseInt(strings.TrimSpace(m[1]), 10, 64); err == nil {
					replyTo = v
				}
			}
			reactions := parseReactions(block)

			out[msgID] = htmlMessage{
				ID:          msgID,
				ReplyToID:   replyTo,
				FromName:    fromName,
				Text:        text,
				PublishedAt: publishedAt.UTC(),
				Reactions:   reactions,
			}
			ordered = append(ordered, msgID)
		}
	}

	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	return out, ordered, nil
}

func cleanHTMLText(s string) string {
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = reTags.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(reSpaces.ReplaceAllString(lines[i], " "))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func parseReactions(block string) map[string]int {
	matches := reReaction.FindAllStringSubmatch(block, -1)
	out := make(map[string]int, len(matches))
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		emoji := strings.TrimSpace(cleanHTMLText(m[1]))
		if emoji == "" {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(m[2]))
		if err != nil || n <= 0 {
			continue
		}
		out["emoji:"+emoji] += n
	}
	return out
}

func loadTGPostMap(ctx context.Context, pool *pgxpool.Pool) (map[int64]tgPostRef, error) {
	rows, err := pool.Query(ctx, `select id, source_post_id, published_at from posts where source='tg'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64]tgPostRef{}
	for rows.Next() {
		var (
			postID       int64
			sourcePostID string
			publishedAt  time.Time
		)
		if err := rows.Scan(&postID, &sourcePostID, &publishedAt); err != nil {
			return nil, err
		}
		msgID, ok := parseTGMessageID(sourcePostID)
		if !ok {
			continue
		}
		out[msgID] = tgPostRef{PostID: postID, PublishedAt: publishedAt.UTC()}
	}
	return out, rows.Err()
}

func resolvePostLink(msgs map[int64]htmlMessage, postByTGMsgID map[int64]tgPostRef, replyToID int64, commentPublishedAt time.Time) (postID int64, parentCommentID *string, ok bool) {
	if _, isPost := postByTGMsgID[replyToID]; !isPost {
		v := strconv.FormatInt(replyToID, 10)
		parentCommentID = &v
	}

	visited := map[int64]struct{}{}
	cur := replyToID
	for cur > 0 {
		if ref, isPost := postByTGMsgID[cur]; isPost {
			// Комментарий не может быть раньше исходного поста.
			if commentPublishedAt.Before(ref.PublishedAt) {
				return 0, nil, false
			}
			return ref.PostID, parentCommentID, true
		}
		if _, seen := visited[cur]; seen {
			return 0, nil, false
		}
		visited[cur] = struct{}{}
		parent, exists := msgs[cur]
		if !exists || parent.ReplyToID == 0 {
			return 0, nil, false
		}
		cur = parent.ReplyToID
	}
	return 0, nil, false
}

func findRootMessage(msgs map[int64]htmlMessage, replyToID int64) (int64, htmlMessage, bool) {
	visited := map[int64]struct{}{}
	cur := replyToID
	for cur > 0 {
		msg, exists := msgs[cur]
		if !exists {
			return 0, htmlMessage{}, false
		}
		if msg.ReplyToID == 0 {
			return cur, msg, true
		}
		if _, seen := visited[cur]; seen {
			return 0, htmlMessage{}, false
		}
		visited[cur] = struct{}{}
		cur = msg.ReplyToID
	}
	return 0, htmlMessage{}, false
}

func upsertComment(ctx context.Context, pool *pgxpool.Pool, postID int64, msg htmlMessage, parentCommentID *string) error {
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var reactions any
	if len(msg.Reactions) > 0 {
		b, err := json.Marshal(msg.Reactions)
		if err != nil {
			return err
		}
		reactions = b
	}

	var text any
	if strings.TrimSpace(msg.Text) != "" {
		text = msg.Text
	}

	_, err := pool.Exec(opCtx, `
insert into post_comments (
  post_id, source, source_comment_id, source_parent_comment_id, author_name, published_at, text, reactions, media, updated_at
) values (
  $1,$2,$3,$4,$5,$6,$7,$8,$9, now()
)
on conflict (source, source_comment_id) do update set
  post_id = excluded.post_id,
  source_parent_comment_id = excluded.source_parent_comment_id,
  author_name = excluded.author_name,
  published_at = excluded.published_at,
  text = excluded.text,
  reactions = excluded.reactions,
  media = excluded.media,
  updated_at = now()
`, postID, "tg", strconv.FormatInt(msg.ID, 10), parentCommentID, nullIfEmpty(msg.FromName), msg.PublishedAt, text, reactions, nil)
	return err
}

func upsertRootPost(ctx context.Context, pool *pgxpool.Pool, msg htmlMessage) (created bool, postID int64, publishedAt time.Time, err error) {
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	publishedAt = msg.PublishedAt.UTC()

	likes := 0
	for _, n := range msg.Reactions {
		likes += n
	}
	var text any
	if strings.TrimSpace(msg.Text) != "" {
		text = msg.Text
	}

	err = pool.QueryRow(opCtx, `
insert into posts (
  source, source_post_id, published_at, text, media, likes_count, views_count, comments_count, updated_at
) values (
  'tg',$1,$2,$3,$4,$5,$6,$7, now()
)
on conflict (source, source_post_id) do nothing
returning id
`, strconv.FormatInt(msg.ID, 10), msg.PublishedAt, text, nil, likes, nil, nil).Scan(&postID)
	if err == nil {
		return true, postID, publishedAt, nil
	}
	if !strings.Contains(err.Error(), "no rows in result set") {
		return false, 0, time.Time{}, err
	}
	err = pool.QueryRow(opCtx, `select id, published_at from posts where source='tg' and source_post_id=$1`, strconv.FormatInt(msg.ID, 10)).Scan(&postID, &publishedAt)
	if err != nil {
		return false, 0, time.Time{}, err
	}
	return false, postID, publishedAt.UTC(), nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func parseTGMessageID(sourcePostID string) (int64, bool) {
	s := strings.TrimSpace(sourcePostID)
	if s == "" {
		return 0, false
	}
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return v, true
	}
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] < '0' || s[i] > '9' {
			if i == len(s)-1 {
				return 0, false
			}
			v, err := strconv.ParseInt(s[i+1:], 10, 64)
			return v, err == nil
		}
	}
	return 0, false
}
