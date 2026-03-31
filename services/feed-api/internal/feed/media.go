package feed

import (
	"encoding/json"
	"path/filepath"
	"strings"

	mediadb "github.com/goritskimihail/mudro/internal/media"
)

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
		u := normalizeMediaURL(item.URL)
		preview := normalizeMediaURL(item.PreviewURL)
		if kind == "" && u == "" && preview == "" && title == "" {
			continue
		}

		out = append(out, feedMediaItem{
			Kind:       kind,
			URL:        u,
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
			return strings.TrimSpace(strings.TrimSpace(v.(string)))
		}
	}
	return ""
}

func normalizeMediaKind(kindRaw, mediaType, mimeType, u, title string) string {
	candidates := []string{kindRaw, mediaType, mimeType, u, title}
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
	lowerURL := strings.ToLower(strings.TrimSpace(u))
	switch {
	case strings.Contains(lowerURL, ".mp4"), strings.Contains(lowerURL, ".webm"):
		return "video"
	case strings.Contains(lowerURL, ".mp3"), strings.Contains(lowerURL, ".m4a"), strings.Contains(lowerURL, ".ogg"):
		return "audio"
	case strings.Contains(lowerURL, ".pdf"):
		return "doc"
	case strings.Contains(lowerURL, ".jpg"), strings.Contains(lowerURL, ".jpeg"), strings.Contains(lowerURL, ".png"), strings.Contains(lowerURL, ".webp"), strings.Contains(lowerURL, ".gif"):
		return "photo"
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
	base := filepath.Base(s)
	if base == "." || base == "/" || base == "" {
		return ""
	}
	return base
}

func normalizeMediaURL(raw string) string {
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
