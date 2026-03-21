package app

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	commentdb "github.com/goritskimihail/mudro/internal/commentmodel"
	"github.com/goritskimihail/mudro/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type csvCommentRow struct {
	Date         time.Time
	MessageID    int64
	ReplyToID    int64
	Sender       string
	SenderID     string
	Message      string
	Reactions    map[string]int
	IsThreadRoot bool
}

type tgPostCandidate struct {
	PostID         int64
	SourcePostID   string
	PublishedAt    time.Time
	NormalizedText string
	LikesCount     int
	ViewsCount     *int
}

var reactionPairRe = regexp.MustCompile(`'([^']+)'\s*:\s*(\d+)`)

func Run() {
	inPath := flag.String("in", "", "path to discussion Messages_*.csv export")
	exportJSONPath := flag.String("export-json", "", "optional path to full Telegram result.json for author names")
	dsn := flag.String("dsn", config.DSN(), "postgres dsn")
	flag.Parse()

	if strings.TrimSpace(*inPath) == "" {
		log.Fatal("flag -in is required")
	}

	rows, err := readCSV(*inPath)
	if err != nil {
		log.Fatalf("read csv: %v", err)
	}
	senderNames, err := loadSenderNames(*exportJSONPath)
	if err != nil {
		log.Fatalf("load sender names: %v", err)
	}
	ordered := orderedRowIDs(rows)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	postCandidates, err := loadTGPostCandidates(ctx, pool)
	if err != nil {
		log.Fatalf("load tg posts: %v", err)
	}

	threadRoots := make(map[int64]tgPostCandidate)
	for _, id := range ordered {
		row := rows[id]
		if !row.IsThreadRoot {
			continue
		}
		ref, ok := matchDiscussionRoot(row, postCandidates)
		if ok {
			threadRoots[row.MessageID] = ref
		}
	}

	imported := 0
	skippedStandalone := 0
	skippedUnmatched := 0
	for _, id := range ordered {
		row := rows[id]
		if row.IsThreadRoot {
			continue
		}
		if row.ReplyToID == 0 {
			skippedStandalone++
			continue
		}
		postID, parentCommentID, ok := resolvePostLink(rows, threadRoots, row.ReplyToID, row.Date)
		if !ok {
			skippedUnmatched++
			continue
		}
		if err := upsertComment(ctx, pool, postID, row, parentCommentID, senderNames); err != nil {
			log.Fatalf("upsert comment %d: %v", row.MessageID, err)
		}
		imported++
	}

	log.Printf(
		"DONE: imported_comments=%d matched_roots=%d skipped_standalone=%d skipped_unmatched=%d csv=%s",
		imported,
		len(threadRoots),
		skippedStandalone,
		skippedUnmatched,
		*inPath,
	)
}

func orderedRowIDs(rows map[int64]csvCommentRow) []int64 {
	out := make([]int64, 0, len(rows))
	for id := range rows {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool {
		left := rows[out[i]]
		right := rows[out[j]]
		if left.Date.Equal(right.Date) {
			return left.MessageID < right.MessageID
		}
		return left.Date.Before(right.Date)
	})
	return out
}

func readCSV(path string) (map[int64]csvCommentRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	index := make(map[string]int, len(header))
	for i, name := range header {
		normalized := strings.TrimSpace(strings.TrimPrefix(name, "\uFEFF"))
		index[normalized] = i
	}
	for _, key := range []string{"date", "message_id", "reply_to_msg_id", "sender", "sender_id", "message", "reactions"} {
		if _, ok := index[key]; !ok {
			return nil, fmt.Errorf("missing required column %q", key)
		}
	}

	rows := make(map[int64]csvCommentRow, 512)
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		messageID, ok := parseInt64(field(record, index, "message_id"))
		if !ok {
			continue
		}
		dateRaw := field(record, index, "date")
		parsedDate, err := parseCSVDate(dateRaw)
		if err != nil {
			return nil, fmt.Errorf("parse date %q: %w", dateRaw, err)
		}

		replyToID, _ := parseInt64(field(record, index, "reply_to_msg_id"))
		senderID := field(record, index, "sender_id")
		row := csvCommentRow{
			Date:         parsedDate.UTC(),
			MessageID:    messageID,
			ReplyToID:    replyToID,
			Sender:       strings.TrimSpace(field(record, index, "sender")),
			SenderID:     strings.TrimSpace(senderID),
			Message:      normalizeCSVMessage(field(record, index, "message")),
			Reactions:    parseCSVReactions(field(record, index, "reactions")),
			IsThreadRoot: replyToID == 0 && strings.TrimSpace(senderID) == "-1",
		}
		rows[messageID] = row
	}

	return rows, nil
}

func loadTGPostCandidates(ctx context.Context, pool *pgxpool.Pool) ([]tgPostCandidate, error) {
	rows, err := pool.Query(ctx, `
select id, source_post_id, published_at, text, likes_count, views_count
from posts
where source = 'tg'
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]tgPostCandidate, 0, 2048)
	for rows.Next() {
		var (
			item       tgPostCandidate
			text       *string
			viewsCount *int
		)
		if err := rows.Scan(&item.PostID, &item.SourcePostID, &item.PublishedAt, &text, &item.LikesCount, &viewsCount); err != nil {
			return nil, err
		}
		item.PublishedAt = item.PublishedAt.UTC()
		item.ViewsCount = viewsCount
		item.NormalizedText = normalizeLookupText(textValue(text))
		out = append(out, item)
	}
	return out, rows.Err()
}

func matchDiscussionRoot(row csvCommentRow, posts []tgPostCandidate) (tgPostCandidate, bool) {
	normalizedText := normalizeLookupText(row.Message)
	bestIdx := -1
	bestScore := int64(1<<62 - 1)

	for i, post := range posts {
		if normalizedText != "" && post.NormalizedText != normalizedText {
			continue
		}
		delta := absDurationSeconds(row.Date.Sub(post.PublishedAt))
		if normalizedText != "" {
			if delta > int64(6*60*60) {
				continue
			}
		} else if delta > int64(30*60) {
			continue
		}
		if bestIdx == -1 || delta < bestScore {
			bestIdx = i
			bestScore = delta
		}
	}

	if bestIdx == -1 {
		return tgPostCandidate{}, false
	}
	return posts[bestIdx], true
}

func resolvePostLink(rows map[int64]csvCommentRow, threadRoots map[int64]tgPostCandidate, replyToID int64, commentPublishedAt time.Time) (postID int64, parentCommentID *string, ok bool) {
	if _, isRoot := threadRoots[replyToID]; !isRoot {
		v := strconv.FormatInt(replyToID, 10)
		parentCommentID = &v
	}

	visited := map[int64]struct{}{replyToID: {}}
	cur := replyToID
	for cur > 0 {
		if ref, isRoot := threadRoots[cur]; isRoot {
			if commentPublishedAt.Before(ref.PublishedAt) {
				return 0, nil, false
			}
			return ref.PostID, parentCommentID, true
		}
		parent, exists := rows[cur]
		if !exists || parent.ReplyToID == 0 {
			return 0, nil, false
		}
		if _, seen := visited[parent.ReplyToID]; seen {
			return 0, nil, false
		}
		visited[parent.ReplyToID] = struct{}{}
		cur = parent.ReplyToID
	}
	return 0, nil, false
}

func upsertComment(ctx context.Context, pool *pgxpool.Pool, postID int64, row csvCommentRow, parentCommentID *string, senderNames map[string]string) error {
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := pool.Begin(opCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(opCtx)

	var text any
	if strings.TrimSpace(row.Message) != "" {
		text = row.Message
	}

	var parentRowID any
	if parentCommentID != nil && strings.TrimSpace(*parentCommentID) != "" {
		var resolvedParentID int64
		err := tx.QueryRow(opCtx, `
select id
from post_comments
where source = $1 and post_id = $2 and source_comment_id = $3
`, "tg", postID, strings.TrimSpace(*parentCommentID)).Scan(&resolvedParentID)
		if err == nil {
			parentRowID = resolvedParentID
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	var commentID int64
	err = tx.QueryRow(opCtx, `
insert into post_comments (
  post_id, source, source_comment_id, source_parent_comment_id, parent_comment_id, author_name, published_at, text, updated_at
) values (
  $1,$2,$3,$4,$5,$6,$7,$8, now()
)
on conflict (source, source_comment_id) do update set
  post_id = excluded.post_id,
  source_parent_comment_id = excluded.source_parent_comment_id,
  parent_comment_id = coalesce(excluded.parent_comment_id, post_comments.parent_comment_id),
  author_name = coalesce(excluded.author_name, post_comments.author_name),
  published_at = excluded.published_at,
  text = coalesce(excluded.text, post_comments.text),
  updated_at = now()
returning id
`, postID, "tg", strconv.FormatInt(row.MessageID, 10), parentCommentID, parentRowID, buildCommentAuthor(row, senderNames), row.Date, text).Scan(&commentID)
	if err != nil {
		return err
	}

	if err := commentdb.SyncCommentReactions(opCtx, tx, commentID, row.Reactions); err != nil {
		return err
	}

	return tx.Commit(opCtx)
}

func field(record []string, index map[string]int, key string) string {
	position, ok := index[key]
	if !ok || position >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[position])
}

func parseCSVDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05-07:00",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, raw); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format")
}

func parseCSVReactions(raw string) map[string]int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]int{}
	}

	out := make(map[string]int)
	for _, match := range reactionPairRe.FindAllStringSubmatch(raw, -1) {
		if len(match) != 3 {
			continue
		}
		count, err := strconv.Atoi(strings.TrimSpace(match[2]))
		if err != nil || count <= 0 {
			continue
		}
		key := normalizeReactionKey(strings.TrimSpace(match[1]))
		out[key] += count
	}
	return out
}

func normalizeReactionKey(raw string) string {
	switch {
	case strings.HasPrefix(raw, "ReactionCustomEmoji("):
		return "custom:" + raw
	case strings.HasPrefix(raw, "ReactionPaid("):
		return "unknown:" + raw
	case strings.HasPrefix(raw, ":") && strings.HasSuffix(raw, ":"):
		return "emoji:" + raw
	default:
		return "emoji:" + raw
	}
}

func normalizeCSVMessage(raw string) string {
	text := strings.TrimSpace(strings.ReplaceAll(raw, "\r\n", "\n"))
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	if text == "" || text == "MediaMessage" {
		return ""
	}
	return text
}

func normalizeLookupText(raw string) string {
	text := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(raw, "\r\n", "\n")))
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return text
}

func buildCommentAuthor(row csvCommentRow, senderNames map[string]string) any {
	if sender := strings.TrimSpace(row.Sender); sender != "" {
		return sender
	}
	senderID := strings.TrimSpace(row.SenderID)
	if senderID == "" || senderID == "-1" {
		return nil
	}
	if sender, ok := senderNames["user"+senderID]; ok && strings.TrimSpace(sender) != "" {
		return sender
	}
	return fmt.Sprintf("Участник #%s", senderID)
}

func loadSenderNames(path string) (map[string]string, error) {
	out := make(map[string]string)
	path = strings.TrimSpace(path)
	if path == "" {
		return out, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var payload struct {
		Messages []struct {
			FromID string `json:"from_id"`
			From   string `json:"from"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(file).Decode(&payload); err != nil {
		return nil, err
	}

	for _, item := range payload.Messages {
		fromID := strings.TrimSpace(item.FromID)
		from := strings.TrimSpace(item.From)
		if fromID == "" || from == "" {
			continue
		}
		if _, exists := out[fromID]; !exists {
			out[fromID] = from
		}
	}

	return out, nil
}

func parseInt64(raw string) (int64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	return v, err == nil
}

func textValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func absDurationSeconds(d time.Duration) int64 {
	if d < 0 {
		d = -d
	}
	return int64(d / time.Second)
}
