package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type csvRow struct {
	Date           time.Time
	MessageID      string
	Message        string
	Views          *int
	Comments       *int
	Reactions      map[string]int
	ReactionsTotal *int
}

var reactionPairRe = regexp.MustCompile(`'([^']+)'\s*:\s*(\d+)`)

func main() {
	inPath := flag.String("in", "", "path to Messages_*.csv export")
	dsn := flag.String("dsn", config.DSN(), "postgres dsn")
	flag.Parse()

	if strings.TrimSpace(*inPath) == "" {
		log.Fatal("flag -in is required")
	}

	rows, err := readCSV(*inPath)
	if err != nil {
		log.Fatalf("read csv: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	imported := 0
	for _, row := range rows {
		if err := upsertTelegramPost(ctx, pool, row); err != nil {
			log.Fatalf("upsert row %s: %v", row.MessageID, err)
		}
		imported++
	}

	log.Printf("DONE: imported=%d csv=%s", imported, *inPath)
}

func readCSV(path string) ([]csvRow, error) {
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
	for _, key := range []string{"date", "message_id", "message", "views", "comments", "reactions", "reactions_total"} {
		if _, ok := index[key]; !ok {
			return nil, fmt.Errorf("missing required column %q", key)
		}
	}

	out := make([]csvRow, 0, 256)
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		dateRaw := field(record, index, "date")
		parsedDate, err := parseCSVDate(dateRaw)
		if err != nil {
			return nil, fmt.Errorf("parse date %q: %w", dateRaw, err)
		}

		messageID := field(record, index, "message_id")
		if messageID == "" {
			continue
		}

		out = append(out, csvRow{
			Date:           parsedDate.UTC(),
			MessageID:      messageID,
			Message:        normalizeCSVMessage(field(record, index, "message")),
			Views:          parseOptionalInt(field(record, index, "views")),
			Comments:       parseOptionalInt(field(record, index, "comments")),
			Reactions:      parseCSVReactions(field(record, index, "reactions")),
			ReactionsTotal: parseOptionalInt(field(record, index, "reactions_total")),
		})
	}

	return out, nil
}

func upsertTelegramPost(ctx context.Context, pool *pgxpool.Pool, row csvRow) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	likesCount := 0
	if row.ReactionsTotal != nil {
		likesCount = *row.ReactionsTotal
	} else {
		for _, count := range row.Reactions {
			likesCount += count
		}
	}

	var (
		postID     int64
		matchedID  *int64
		messageAny = anyText(row.Message)
	)
	if matched, err := findExistingLogicalPost(ctx, tx, row); err != nil {
		return err
	} else {
		matchedID = matched
	}

	if matchedID != nil {
		err = tx.QueryRow(ctx, `
update posts
set published_at = $2,
    text = coalesce($3, text),
    likes_count = $4,
    views_count = $5,
    comments_count = $6,
    updated_at = now()
where id = $1
returning id
`, *matchedID, row.Date, messageAny, likesCount, row.Views, row.Comments).Scan(&postID)
	} else {
		err = tx.QueryRow(ctx, `
insert into posts (
  source, source_post_id,
  published_at, text,
  likes_count, views_count, comments_count,
  updated_at
) values (
  'tg', $1,
  $2, $3,
  $4, $5, $6,
  now()
)
on conflict (source, source_post_id) do update set
  published_at = excluded.published_at,
  text = coalesce(excluded.text, posts.text),
  likes_count = excluded.likes_count,
  views_count = excluded.views_count,
  comments_count = excluded.comments_count,
  updated_at = now()
returning id
`, row.MessageID, row.Date, messageAny, likesCount, row.Views, row.Comments).Scan(&postID)
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `delete from post_reactions where post_id = $1`, postID); err != nil {
		return err
	}
	for raw, count := range row.Reactions {
		if count <= 0 {
			continue
		}
		if _, err := tx.Exec(ctx, `
insert into post_reactions (post_id, emoji, count)
values ($1, $2, $3)
`, postID, raw, count); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func findExistingLogicalPost(ctx context.Context, tx pgx.Tx, row csvRow) (*int64, error) {
	if strings.TrimSpace(row.Message) == "" {
		return nil, nil
	}

	var matchedID int64
	err := tx.QueryRow(ctx, `
select id
from posts
where source = 'tg'
  and source_post_id <> $1
  and coalesce(text, '') = $2
  and abs(extract(epoch from (published_at - $3))) <= 21600
order by coalesce(comments_count, 0) desc, published_at desc, id desc
limit 1
`, row.MessageID, strings.TrimSpace(row.Message), row.Date).Scan(&matchedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &matchedID, nil
}

func field(record []string, index map[string]int, key string) string {
	position, ok := index[key]
	if !ok || position >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[position])
}

func parseOptionalInt(raw string) *int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &value
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

func anyText(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return raw
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
