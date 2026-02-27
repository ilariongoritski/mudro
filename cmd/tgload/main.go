package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type feedItem struct {
	Source       string      `json:"source"`
	SourcePostID string      `json:"source_post_id"`
	PublishedAt  string      `json:"published_at"`
	Text         string      `json:"text"`
	Stats        feedStats   `json:"stats"`
	Media        []mediaItem `json:"media"`
	AudioTracks  []string    `json:"audio_tracks"`
}

type feedStats struct {
	Views     *int           `json:"views"`
	Likes     *int           `json:"likes"`
	Comments  *int           `json:"comments"`
	Reactions map[string]int `json:"reactions"`
}

type mediaItem struct {
	Kind       string         `json:"kind"`
	URL        string         `json:"url"`
	PreviewURL string         `json:"preview_url"`
	Width      *int           `json:"width"`
	Height     *int           `json:"height"`
	Position   int            `json:"position"`
	Title      string         `json:"title"`
	Extra      map[string]any `json:"extra"`
}

func main() {
	inPath := flag.String("in", "feed_items.json", "path to feed_items JSON")
	dsn := flag.String("dsn", config.DSN(), "postgres dsn")
	flag.Parse()

	raw, err := os.ReadFile(*inPath)
	if err != nil {
		log.Fatalf("read input: %v", err)
	}

	var items []feedItem
	if err := json.Unmarshal(raw, &items); err != nil {
		log.Fatalf("parse input: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	inserted := 0
	for i := range items {
		if strings.ToLower(items[i].Source) != "tg" {
			continue
		}
		if err := upsertTGPost(context.Background(), pool, items[i]); err != nil {
			log.Fatalf("import item[%d] source_post_id=%s: %v", i, items[i].SourcePostID, err)
		}
		inserted++
	}

	log.Printf("DONE: imported tg posts=%d", inserted)
}

func upsertTGPost(ctx context.Context, pool *pgxpool.Pool, it feedItem) error {
	txCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := pool.Begin(txCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(txCtx)

	publishedAt, err := time.Parse(time.RFC3339, it.PublishedAt)
	if err != nil {
		return fmt.Errorf("parse published_at: %w", err)
	}

	likes := 0
	if it.Stats.Likes != nil {
		likes = *it.Stats.Likes
	}

	mediaJSON, err := buildMediaJSON(it.Media, it.AudioTracks)
	if err != nil {
		return fmt.Errorf("build media: %w", err)
	}

	var postID int64
	if err := tx.QueryRow(txCtx, `
insert into posts (
  source, source_post_id, published_at, text, media, likes_count, views_count, comments_count, updated_at
) values (
  $1,$2,$3,$4,$5,$6,$7,$8, now()
)
on conflict (source, source_post_id) do update set
  published_at = excluded.published_at,
  text = excluded.text,
  media = excluded.media,
  likes_count = excluded.likes_count,
  views_count = excluded.views_count,
  comments_count = excluded.comments_count,
  updated_at = now()
returning id
`,
		"tg",
		it.SourcePostID,
		publishedAt,
		nullIfEmpty(it.Text),
		mediaJSON,
		likes,
		it.Stats.Views,
		it.Stats.Comments,
	).Scan(&postID); err != nil {
		return err
	}

	if _, err := tx.Exec(txCtx, `delete from post_reactions where post_id = $1`, postID); err != nil {
		return err
	}

	keys := make([]string, 0, len(it.Stats.Reactions))
	for k := range it.Stats.Reactions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, emoji := range keys {
		count := it.Stats.Reactions[emoji]
		if count <= 0 {
			continue
		}
		if _, err := tx.Exec(txCtx, `
insert into post_reactions (post_id, emoji, count)
values ($1,$2,$3)
on conflict (post_id, emoji) do update set count = excluded.count
`, postID, emoji, count); err != nil {
			return err
		}
	}

	return tx.Commit(txCtx)
}

func buildMediaJSON(media []mediaItem, tracks []string) ([]byte, error) {
	out := make([]map[string]any, 0, len(media)+len(tracks))
	for _, m := range media {
		row := map[string]any{
			"kind":     strings.TrimSpace(m.Kind),
			"url":      strings.TrimSpace(m.URL),
			"position": m.Position,
		}
		if strings.TrimSpace(m.Title) != "" {
			row["title"] = strings.TrimSpace(m.Title)
		}
		if strings.TrimSpace(m.PreviewURL) != "" {
			row["preview_url"] = strings.TrimSpace(m.PreviewURL)
		}
		if m.Width != nil && *m.Width > 0 {
			row["width"] = *m.Width
		}
		if m.Height != nil && *m.Height > 0 {
			row["height"] = *m.Height
		}
		if len(m.Extra) > 0 {
			row["extra"] = m.Extra
		}
		out = append(out, row)
	}

	pos := len(out)
	for _, t := range tracks {
		track := strings.TrimSpace(t)
		if track == "" {
			continue
		}
		out = append(out, map[string]any{
			"kind":     "audio",
			"title":    track,
			"position": pos,
		})
		pos++
	}

	if len(out) == 0 {
		return nil, nil
	}
	return json.Marshal(out)
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
