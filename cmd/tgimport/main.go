package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ===== Input (Telegram export) =====

type Export struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	ID       int64       `json:"id"`
	Messages []TGMessage `json:"messages"`
}

type TGMessage struct {
	ID           int64        `json:"id"`
	Type         string       `json:"type"` // "message" | "service"
	Date         string       `json:"date"`
	DateUnixtime string       `json:"date_unixtime"`
	Text         any          `json:"text"` // string | []any
	TextEntities []TGEntity   `json:"text_entities"`
	Reactions    []TGReaction `json:"reactions"`

	// media (may be absent)
	Photo         string `json:"photo"`
	PhotoFileSize *int64 `json:"photo_file_size"`

	MediaType    string `json:"media_type"` // audio_file, sticker, video_message, ...
	File         string `json:"file"`       // local path or "(File not included...)"
	FileName     string `json:"file_name"`  // for audio etc (null -> "")
	FileSize     *int64 `json:"file_size"`
	Thumbnail    string `json:"thumbnail"` // may be "(File not included...)"
	MimeType     string `json:"mime_type"`
	DurationSec  *int   `json:"duration_seconds"`
	Width        *int   `json:"width"`
	Height       *int   `json:"height"`
	Performer    string `json:"performer"`
	Title        string `json:"title"`
	StickerEmoji string `json:"sticker_emoji"`
}

type TGEntity struct {
	Type string `json:"type"` // plain, link, text_link, bold, ...
	Text string `json:"text"`
	Href string `json:"href"` // for text_link
}

type TGReaction struct {
	Type       string `json:"type"` // emoji | custom_emoji
	Count      int    `json:"count"`
	Emoji      string `json:"emoji"`
	DocumentID string `json:"document_id"` // in your export it's a local sticker path
}

// ===== Output (your feed contract) =====

type FeedItem struct {
	ID           string      `json:"id"`
	Source       string      `json:"source"` // "tg"
	SourcePostID string      `json:"source_post_id"`
	PublishedAt  string      `json:"published_at"`
	CollectedAt  string      `json:"collected_at"`
	URL          *string     `json:"url"`
	Text         string      `json:"text"`
	Stats        Stats       `json:"stats"`
	Media        []MediaItem `json:"media"`
	AudioTracks  []string    `json:"audio_tracks"`
}

type Stats struct {
	Views     *int           `json:"views"`
	Likes     *int           `json:"likes"`
	Comments  *int           `json:"comments"`
	Reactions map[string]int `json:"reactions"`
}

type MediaItem struct {
	Kind       string         `json:"kind"`
	URL        string         `json:"url"` // пока локальный путь из экспорта
	PreviewURL *string        `json:"preview_url"`
	Width      *int           `json:"width"`
	Height     *int           `json:"height"`
	Position   int            `json:"position"`
	Extra      map[string]any `json:"extra"` // nil -> null
}

func main() {
	inPath := flag.String("in", "result.json", "path to Telegram export JSON")
	outPath := flag.String("out", "feed_items.json", "output JSON path")
	collectedAtFlag := flag.String("collected-at", "", "RFC3339 timestamp (default: now UTC)")
	pretty := flag.Bool("pretty", true, "pretty-print JSON")
	flag.Parse()

	collectedAt := *collectedAtFlag
	if collectedAt == "" {
		collectedAt = time.Now().UTC().Format(time.RFC3339)
	} else {
		if _, err := time.Parse(time.RFC3339, collectedAt); err != nil {
			die("invalid -collected-at (need RFC3339, e.g. 2026-02-23T10:00:00Z): %v", err)
		}
	}

	b, err := os.ReadFile(*inPath)
	if err != nil {
		die("read input: %v", err)
	}

	var exp Export
	if err := json.Unmarshal(b, &exp); err != nil {
		die("parse input json: %v", err)
	}

	out := make([]FeedItem, 0, len(exp.Messages))

	for _, m := range exp.Messages {
		// 1) service не нужны
		if m.Type != "message" {
			continue
		}

		publishedAt, ok := toUTC(m.DateUnixtime)
		if !ok {
			continue
		}

		text := buildText(m.TextEntities, m.Text)
		likes, reactions := buildReactions(m.Reactions)

		media, audio := buildMediaAndAudio(m)

		fi := FeedItem{
			ID:           fmt.Sprintf("tg:%d", m.ID),
			Source:       "tg",
			SourcePostID: strconv.FormatInt(m.ID, 10),
			PublishedAt:  publishedAt,
			CollectedAt:  collectedAt,
			URL:          nil, // позже: https://t.me/<handle>/<id> если знаешь handle
			Text:         text,
			Stats: Stats{
				Views:     nil,
				Likes:     likes,
				Comments:  nil,
				Reactions: reactions,
			},
			Media:       media,
			AudioTracks: audio,
		}

		// гарантируем [] вместо null
		if fi.Media == nil {
			fi.Media = []MediaItem{}
		}
		if fi.AudioTracks == nil {
			fi.AudioTracks = []string{}
		}

		out = append(out, fi)
	}

	// write
	if dir := filepath.Dir(*outPath); dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}

	f, err := os.Create(*outPath)
	if err != nil {
		die("create output: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if *pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(out); err != nil {
		die("write output: %v", err)
	}

	fmt.Printf("OK: wrote %d items to %s (collected_at=%s)\n", len(out), *outPath, collectedAt)
}

func toUTC(unixStr string) (string, bool) {
	sec, err := strconv.ParseInt(unixStr, 10, 64)
	if err != nil {
		return "", false
	}
	return time.Unix(sec, 0).UTC().Format(time.RFC3339), true
}

func buildText(ents []TGEntity, raw any) string {
	// text_entities есть даже когда raw text — массив сегментов.
	if len(ents) > 0 {
		var b strings.Builder
		for _, e := range ents {
			b.WriteString(e.Text)
		}
		return b.String()
	}

	// fallback: raw может быть string или []any
	if s, ok := raw.(string); ok {
		return s
	}
	arr, ok := raw.([]any)
	if !ok {
		return ""
	}
	var b strings.Builder
	for _, seg := range arr {
		if s, ok := seg.(string); ok {
			b.WriteString(s)
			continue
		}
		if m, ok := seg.(map[string]any); ok {
			if txt, ok := m["text"].(string); ok {
				b.WriteString(txt)
			}
		}
	}
	return b.String()
}

func buildReactions(rs []TGReaction) (*int, map[string]int) {
	sum := 0
	m := make(map[string]int) // даже если пусто — будет {}

	for _, r := range rs {
		sum += r.Count

		var key string
		switch r.Type {
		case "emoji":
			key = "emoji:" + r.Emoji
		case "custom_emoji":
			key = "custom:" + r.DocumentID
		default:
			key = "unknown:"
		}

		m[key] += r.Count
	}

	return &sum, m // likes всегда число (0 тоже ок)
}

func buildMediaAndAudio(m TGMessage) ([]MediaItem, []string) {
	media := make([]MediaItem, 0, 4)
	audio := make([]string, 0, 1)
	pos := 1

	// 1) main media: photo
	if m.Photo != "" {
		media = append(media, MediaItem{
			Kind:       "photo",
			URL:        m.Photo, // локальный путь из экспорта
			PreviewURL: nil,
			Width:      m.Width,
			Height:     m.Height,
			Position:   pos,
			Extra: map[string]any{
				"file_size": m.PhotoFileSize,
			},
		})
		pos++
	}

	// 2) main media: file/video/sticker/audio etc
	if m.MediaType != "" || m.File != "" {
		kind := "file"
		switch m.MediaType {
		case "video_message", "video_file":
			kind = "video"
		case "animation", "sticker":
			kind = "gif"
		case "audio_file", "voice_message":
			kind = "file"
		default:
			kind = "file"
		}

		url := m.File
		missing := false
		// если экспорт не скачал файл — будет строка "(File not included...)"
		if url == "" || strings.HasPrefix(url, "(") {
			missing = true
			if m.FileName != "" {
				url = "missing://" + m.FileName
			} else if m.MediaType != "" {
				url = "missing://" + m.MediaType
			} else {
				url = "missing://unknown"
			}
		}

		var prev *string
		if m.Thumbnail != "" && !strings.HasPrefix(m.Thumbnail, "(") {
			prev = &m.Thumbnail
		}

		extra := map[string]any{
			"media_type":    m.MediaType,
			"mime_type":     m.MimeType,
			"duration_sec":  m.DurationSec,
			"file_size":     m.FileSize,
			"file_name":     m.FileName,
			"missing":       missing,
			"sticker_emoji": m.StickerEmoji,
		}

		media = append(media, MediaItem{
			Kind:       kind,
			URL:        url,
			PreviewURL: prev,
			Width:      m.Width,
			Height:     m.Height,
			Position:   pos,
			Extra:      extra,
		})
		pos++

		// audio_tracks: performer — title (если есть)
		if m.MediaType == "audio_file" {
			track := strings.TrimSpace(strings.Join([]string{m.Performer, m.Title}, " — "))
			track = strings.Trim(track, "— ")
			if track == "" {
				track = m.FileName
			}
			if strings.TrimSpace(track) != "" {
				audio = append(audio, track)
			}
		}
	}

	// 3) links from text_entities as MediaItem(kind=link)
	seen := map[string]bool{}
	for _, e := range m.TextEntities {
		var u string
		switch e.Type {
		case "link":
			u = e.Text
		case "text_link":
			if e.Href != "" {
				u = e.Href
			} else {
				u = e.Text
			}
		default:
			continue
		}
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true

		media = append(media, MediaItem{
			Kind:       "link",
			URL:        u,
			PreviewURL: nil,
			Width:      nil,
			Height:     nil,
			Position:   pos,
			Extra:      nil,
		})
		pos++
	}

	return media, audio
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
