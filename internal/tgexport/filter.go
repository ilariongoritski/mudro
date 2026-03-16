package tgexport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type Export struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	ID               int64  `json:"id"`
	Type             string `json:"type"`
	From             string `json:"from"`
	FromID           string `json:"from_id"`
	ReplyToMessageID int64  `json:"reply_to_message_id"`
}

func LooksLikeMudroAuthor(from, fromID string) bool {
	name := normalizeAuthor(from)
	if strings.Contains(name, "мудро") || strings.Contains(name, "mudro") {
		return true
	}
	id := strings.ToLower(strings.TrimSpace(fromID))
	if strings.HasPrefix(id, "channel") && (strings.Contains(name, "1001") || strings.Contains(name, "100.1")) {
		return true
	}
	return false
}

func IsVisiblePost(msg Message) bool {
	return msg.Type == "message" && msg.ReplyToMessageID == 0 && LooksLikeMudroAuthor(msg.From, msg.FromID)
}

func LoadVisibleSourcePostIDs(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var export Export
	if err := json.Unmarshal(raw, &export); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	ids := make([]string, 0, len(export.Messages))
	for _, msg := range export.Messages {
		if !IsVisiblePost(msg) {
			continue
		}
		ids = append(ids, fmt.Sprintf("%d", msg.ID))
	}
	sort.Strings(ids)
	return ids, nil
}

func LoadVisibleSourcePostIDsFromRepo(repoRoot string) ([]string, string, error) {
	candidates := []string{
		filepath.Join(repoRoot, "data", "nu", "result.json"),
		filepath.Join(repoRoot, "data", "tg-export", "result.json"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		ids, err := LoadVisibleSourcePostIDs(candidate)
		if err != nil {
			return nil, candidate, err
		}
		return ids, candidate, nil
	}

	return nil, "", fmt.Errorf("telegram export result.json not found under %s", repoRoot)
}

func normalizeAuthor(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	var out []rune
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			out = append(out, r)
		case unicode.IsSpace(r):
			continue
		case r == '.' || r == '_' || r == '-':
			continue
		}
	}
	return string(out)
}
