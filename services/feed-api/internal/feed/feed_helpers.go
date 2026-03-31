package feed

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/goritskimihail/mudro/internal/posts"
)

func compactText(s *string) string {
	if s == nil {
		return ""
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return ""
	}
	return t
}

func buildFeedReactions(reactions map[string]int) []feedReaction {
	if len(reactions) == 0 {
		return nil
	}
	keys := make([]string, 0, len(reactions))
	for k, v := range reactions {
		if v > 0 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		vi := reactions[keys[i]]
		vj := reactions[keys[j]]
		if vi == vj {
			return keys[i] < keys[j]
		}
		return vi > vj
	})
	out := make([]feedReaction, 0, len(keys))
	for _, k := range keys {
		out = append(out, feedReaction{
			Label: reactionLabel(k),
			Count: reactions[k],
			Raw:   k,
		})
	}
	return out
}

func reactionLabel(raw string) string {
	switch {
	case strings.HasPrefix(raw, "emoji:"):
		label := strings.TrimSpace(strings.TrimPrefix(raw, "emoji:"))
		if label != "" {
			return label
		}
	case strings.HasPrefix(raw, "custom:"):
		return "✨"
	case strings.HasPrefix(raw, "unknown:"):
		return "?"
	}
	if raw == "" {
		return "?"
	}
	return raw
}

func buildFeedComments(comments []posts.Comment) []feedComment {
	if len(comments) == 0 {
		return nil
	}
	out := make([]feedComment, 0, len(comments))
	for _, c := range comments {
		out = append(out, feedComment{
			SourceCommentID: c.SourceCommentID,
			ParentCommentID: c.ParentCommentID,
			AuthorName:      c.AuthorName,
			PublishedAt:     c.PublishedAt,
			Text:            c.Text,
			Reactions:       buildFeedReactions(c.Reactions),
			Media:           buildFeedMedia(c.Media),
		})
	}
	return out
}

func buildFeedMedia(media json.RawMessage) []feedMediaItem {
	if len(media) == 0 {
		return nil
	}
	items := posts.ParseMediaItems(media)
	out := make([]feedMediaItem, 0, len(items))
	for _, item := range items {
		out = append(out, feedMediaItem{
			Kind:       item.Kind,
			URL:        item.URL,
			PreviewURL: item.PreviewURL,
			Title:      item.Title,
			Width:      item.Width,
			Height:     item.Height,
			Position:   item.Position,
			IsImage:    item.Kind == "photo" || item.Kind == "gif" || item.Kind == "image",
			IsAudio:    item.Kind == "audio",
			IsVideo:    item.Kind == "video",
			IsDocument: item.Kind == "doc",
			IsLink:     item.Kind == "link",
		})
	}
	return out
}

func sourceTotals(stats []posts.SourceStat) (vkTotal, tgTotal int64) {
	for _, st := range stats {
		switch st.Source {
		case "vk":
			vkTotal = st.Posts
		case "tg":
			tgTotal = st.Posts
		}
	}
	return vkTotal, tgTotal
}
