package main

import (
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

	"github.com/goritskimihail/mudro/internal/tgexport"
)

type feedItem struct {
	ID           string      `json:"id"`
	Source       string      `json:"source"`
	SourcePostID string      `json:"source_post_id"`
	PublishedAt  string      `json:"published_at"`
	CollectedAt  string      `json:"collected_at"`
	URL          *string     `json:"url"`
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
	PreviewURL *string        `json:"preview_url"`
	Width      *int           `json:"width"`
	Height     *int           `json:"height"`
	Position   int            `json:"position"`
	Extra      map[string]any `json:"extra"`
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
	inDir := flag.String("dir", "data/tg-export", "directory with messages*.html")
	outPath := flag.String("out", "data/tg-export/feed_items.json", "output JSON path")
	collectedAtFlag := flag.String("collected-at", "", "RFC3339 timestamp (default: now UTC)")
	pretty := flag.Bool("pretty", true, "pretty-print JSON")
	flag.Parse()

	collectedAt := *collectedAtFlag
	if collectedAt == "" {
		collectedAt = time.Now().UTC().Format(time.RFC3339)
	} else if _, err := time.Parse(time.RFC3339, collectedAt); err != nil {
		log.Fatalf("invalid -collected-at: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(*inDir, "messages*.html"))
	if err != nil {
		log.Fatalf("glob html files: %v", err)
	}
	if len(files) == 0 {
		log.Fatalf("no files matched: %s", filepath.Join(*inDir, "messages*.html"))
	}
	sort.Strings(files)

	items := make([]feedItem, 0, 2048)
	for _, f := range files {
		part, err := parseHTMLFile(f, collectedAt)
		if err != nil {
			log.Fatalf("parse %s: %v", f, err)
		}
		items = append(items, part...)
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].SourcePostID < items[j].SourcePostID
	})

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil && filepath.Dir(*outPath) != "." {
		log.Fatalf("mkdir output dir: %v", err)
	}
	f, err := os.Create(*outPath)
	if err != nil {
		log.Fatalf("create output: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if *pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(items); err != nil {
		log.Fatalf("write json: %v", err)
	}

	log.Printf("OK: wrote %d items to %s", len(items), *outPath)
}

func parseHTMLFile(path, collectedAt string) ([]feedItem, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := string(b)
	starts := reMsgStart.FindAllStringSubmatchIndex(s, -1)
	if len(starts) == 0 {
		return nil, nil
	}

	out := make([]feedItem, 0, len(starts))
	lastFromName := ""
	for i, st := range starts {
		blockStart := st[0]
		blockEnd := len(s)
		if i+1 < len(starts) {
			blockEnd = starts[i+1][0]
		}
		block := s[blockStart:blockEnd]
		msgID := s[st[2]:st[3]]

		dateMatch := reDateTitle.FindStringSubmatch(block)
		if len(dateMatch) < 2 {
			continue
		}
		publishedAt, err := parseTGTitleDate(dateMatch[1])
		if err != nil {
			continue
		}

		text := ""
		if m := reTextBlock.FindStringSubmatch(block); len(m) >= 2 {
			text = cleanHTMLText(m[1])
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

		replyTo := int64(0)
		if m := reReplyTo.FindStringSubmatch(block); len(m) >= 2 {
			if v, err := strconv.ParseInt(strings.TrimSpace(m[1]), 10, 64); err == nil {
				replyTo = v
			}
		}

		if replyTo != 0 || !tgexport.LooksLikeMudroAuthor(fromName, "") {
			continue
		}

		likes, reactions := parseReactions(block)

		out = append(out, feedItem{
			ID:           "tg:" + msgID,
			Source:       "tg",
			SourcePostID: msgID,
			PublishedAt:  publishedAt,
			CollectedAt:  collectedAt,
			URL:          nil,
			Text:         text,
			Stats: feedStats{
				Views:     nil,
				Likes:     likes,
				Comments:  nil,
				Reactions: reactions,
			},
			Media:       []mediaItem{},
			AudioTracks: []string{},
		})
	}
	return out, nil
}

func parseTGTitleDate(s string) (string, error) {
	// example: 04.11.2021 12:51:35 UTC+03:00
	t, err := time.Parse("02.01.2006 15:04:05 UTC-07:00", strings.TrimSpace(s))
	if err != nil {
		return "", err
	}
	return t.UTC().Format(time.RFC3339), nil
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

func parseReactions(block string) (*int, map[string]int) {
	matches := reReaction.FindAllStringSubmatch(block, -1)
	reactions := make(map[string]int, len(matches))
	sum := 0
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
		reactions["emoji:"+emoji] += n
		sum += n
	}
	if len(reactions) == 0 {
		reactions = map[string]int{}
	}
	likes := sum
	return &likes, reactions
}
