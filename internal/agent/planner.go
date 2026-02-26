package agent

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var todoLineRe = regexp.MustCompile(`^- \[ \] .+`)

func PlanFromTodo(ctx context.Context, repoRoot string, q *Repository) (int, error) {
	path := filepath.Join(repoRoot, ".codex", "todo.md")
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open todo: %w", err)
	}
	defer f.Close()

	enqueued := 0
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if !todoLineRe.MatchString(line) {
			continue
		}

		text := strings.TrimPrefix(line, "- [ ] ")
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		h := sha1.Sum([]byte(strings.ToLower(text)))
		dedupeKey := "todo:" + hex.EncodeToString(h[:])

		payload := map[string]any{
			"source": "todo.md",
			"text":   text,
		}
		id, err := q.Enqueue(ctx, "todo_item", payload, 10, time.Now(), 3, dedupeKey)
		if err != nil {
			return enqueued, err
		}
		if id > 0 {
			enqueued++
		}
	}
	if err := s.Err(); err != nil {
		return enqueued, fmt.Errorf("scan todo: %w", err)
	}
	return enqueued, nil
}
