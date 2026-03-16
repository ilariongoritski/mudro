package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	mediadb "github.com/goritskimihail/mudro/internal/media"
	"github.com/jackc/pgx/v5/pgxpool"
)

type exportFile struct {
	Messages []tgMessage `json:"messages"`
}

type tgMessage struct {
	ID            int64      `json:"id"`
	Type          string     `json:"type"`
	ReplyToID     int64      `json:"reply_to_message_id"`
	Photo         string     `json:"photo"`
	PhotoFileSize *int64     `json:"photo_file_size"`
	MediaType     string     `json:"media_type"`
	File          string     `json:"file"`
	FileName      string     `json:"file_name"`
	FileSize      *int64     `json:"file_size"`
	Thumbnail     string     `json:"thumbnail"`
	MimeType      string     `json:"mime_type"`
	DurationSec   *int       `json:"duration_seconds"`
	Width         *int       `json:"width"`
	Height        *int       `json:"height"`
	StickerEmoji  string     `json:"sticker_emoji"`
	TextEntities  []tgEntity `json:"text_entities"`
}

type tgEntity struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Href string `json:"href"`
}

type commentState struct {
	ID       int64
	HasMedia bool
}

func main() {
	dir := flag.String("dir", "data/nu", "directory with result.json")
	dsn := flag.String("dsn", config.DSN(), "postgres dsn")
	overwrite := flag.Bool("overwrite", false, "replace existing comment media if present")
	flag.Parse()

	resultPath := filepath.Join(*dir, "result.json")
	raw, err := os.ReadFile(resultPath)
	if err != nil {
		log.Fatalf("read result.json: %v", err)
	}

	var exp exportFile
	if err := json.Unmarshal(raw, &exp); err != nil {
		log.Fatalf("parse result.json: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	states, err := loadCommentStates(ctx, pool)
	if err != nil {
		log.Fatalf("load comment states: %v", err)
	}

	scanned := 0
	replyMedia := 0
	matched := 0
	updated := 0
	skippedNoMedia := 0
	skippedMissing := 0
	skippedExisting := 0

	for _, msg := range exp.Messages {
		if msg.Type != "message" {
			continue
		}
		if msg.ReplyToID == 0 {
			continue
		}
		scanned++
		media := buildMedia(msg)
		if len(media) == 0 {
			skippedNoMedia++
			continue
		}
		replyMedia++

		state, ok := states[msg.ID]
		if !ok {
			skippedMissing++
			continue
		}
		matched++
		if state.HasMedia && !*overwrite {
			skippedExisting++
			continue
		}
		if err := syncCommentMedia(ctx, pool, state.ID, media); err != nil {
			log.Fatalf("sync comment media msg=%d: %v", msg.ID, err)
		}
		updated++
	}

	log.Printf("DONE: scanned_reply=%d reply_with_media=%d matched_existing=%d updated=%d skipped_no_media=%d skipped_missing=%d skipped_existing=%d dir=%s overwrite=%v", scanned, replyMedia, matched, updated, skippedNoMedia, skippedMissing, skippedExisting, *dir, *overwrite)
}

func loadCommentStates(ctx context.Context, pool *pgxpool.Pool) (map[int64]commentState, error) {
	rows, err := pool.Query(ctx, `
select
  pc.id,
  pc.source_comment_id,
  exists (select 1 from comment_media_links cml where cml.comment_id = pc.id) as has_media,
  coalesce(jsonb_array_length(pc.media), 0) > 0 as has_legacy_media
from post_comments pc
where pc.source = 'tg'
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64]commentState, 2048)
	for rows.Next() {
		var (
			id              int64
			sourceCommentID string
			hasMedia        bool
			hasLegacy       bool
		)
		if err := rows.Scan(&id, &sourceCommentID, &hasMedia, &hasLegacy); err != nil {
			return nil, err
		}
		msgID, ok := parseInt64(sourceCommentID)
		if !ok {
			continue
		}
		out[msgID] = commentState{ID: id, HasMedia: hasMedia || hasLegacy}
	}
	return out, rows.Err()
}

func syncCommentMedia(ctx context.Context, pool *pgxpool.Pool, commentID int64, media []mediadb.Item) error {
	opCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := pool.Begin(opCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(opCtx)

	encoded, err := json.Marshal(media)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(opCtx, `update post_comments set media = $2, updated_at = now() where id = $1`, commentID, encoded); err != nil {
		return err
	}
	if err := mediadb.SyncCommentLinks(opCtx, tx, commentID, "tg", encoded); err != nil {
		return err
	}
	return tx.Commit(opCtx)
}

func buildMedia(m tgMessage) []mediadb.Item {
	media := make([]mediadb.Item, 0, 4)
	pos := 1

	if strings.TrimSpace(m.Photo) != "" {
		media = append(media, mediadb.Item{
			Kind:     "photo",
			URL:      strings.TrimSpace(m.Photo),
			Width:    derefInt(m.Width),
			Height:   derefInt(m.Height),
			Position: pos,
			Extra: map[string]any{
				"file_size": m.PhotoFileSize,
			},
		})
		pos++
	}

	if strings.TrimSpace(m.MediaType) != "" || strings.TrimSpace(m.File) != "" {
		kind := "file"
		switch strings.TrimSpace(m.MediaType) {
		case "video_message", "video_file":
			kind = "video"
		case "animation", "sticker":
			kind = "gif"
		case "audio_file", "voice_message":
			kind = "file"
		}

		url := strings.TrimSpace(m.File)
		missing := false
		if url == "" || strings.HasPrefix(url, "(") {
			missing = true
			switch {
			case strings.TrimSpace(m.FileName) != "":
				url = "missing://" + strings.TrimSpace(m.FileName)
			case strings.TrimSpace(m.MediaType) != "":
				url = "missing://" + strings.TrimSpace(m.MediaType)
			default:
				url = "missing://unknown"
			}
		}

		media = append(media, mediadb.Item{
			Kind:       kind,
			URL:        url,
			PreviewURL: trimMissingPreview(strings.TrimSpace(m.Thumbnail)),
			Width:      derefInt(m.Width),
			Height:     derefInt(m.Height),
			Position:   pos,
			Extra: map[string]any{
				"media_type":    m.MediaType,
				"mime_type":     m.MimeType,
				"duration_sec":  m.DurationSec,
				"file_size":     m.FileSize,
				"file_name":     m.FileName,
				"missing":       missing,
				"sticker_emoji": m.StickerEmoji,
			},
		})
		pos++
	}

	seenLinks := map[string]bool{}
	for _, entity := range m.TextEntities {
		var href string
		switch entity.Type {
		case "link":
			href = entity.Text
		case "text_link":
			if entity.Href != "" {
				href = entity.Href
			} else {
				href = entity.Text
			}
		default:
			continue
		}
		href = strings.TrimSpace(href)
		if href == "" || seenLinks[href] {
			continue
		}
		seenLinks[href] = true
		media = append(media, mediadb.Item{Kind: "link", URL: href, Position: pos})
		pos++
	}

	return media
}

func trimMissingPreview(v string) string {
	if v == "" || strings.HasPrefix(v, "(") {
		return ""
	}
	return v
}

func parseInt64(v string) (int64, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	return n, err == nil
}

func derefInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
