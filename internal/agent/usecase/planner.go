package usecase

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
var riskyTodoRe = regexp.MustCompile(`(?i)\b(drop|truncate|reset|rm\s+-rf|docker compose down -v|alter table|delete from)\b`)

func PlanFromTodo(ctx context.Context, repoRoot string, q TaskUsecase) (int, error) {
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
		var (
			id  int64
			err error
		)
		kind := detectTaskKind(text)
		if isRiskyTodo(text) {
			id, err = q.EnqueueWaitingApproval(ctx, kind, payload, 10, time.Now(), 3, dedupeKey)
		} else {
			id, err = q.Enqueue(ctx, kind, payload, 10, time.Now(), 3, dedupeKey)
		}
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

func isRiskyTodo(text string) bool {
	return riskyTodoRe.MatchString(strings.TrimSpace(text))
}

func detectTaskKind(text string) string {
	t := strings.ToLower(strings.TrimSpace(text))
	switch {
	case strings.Contains(t, "dbcheck"),
		strings.Contains(t, "select 1"),
		strings.Contains(t, "проверка бд"),
		strings.Contains(t, "проверки бд"),
		strings.Contains(t, "make dbcheck"):
		return "db_check"
	case strings.Contains(t, `\dt`),
		strings.Contains(t, "список таблиц"),
		strings.Contains(t, "make tables"):
		return "tables_check"
	case strings.Contains(t, "count(*) from posts"),
		strings.Contains(t, "count-posts"),
		strings.Contains(t, "количество постов"),
		strings.Contains(t, "make count-posts"):
		return "count_posts"
	case strings.Contains(t, "make test"),
		strings.Contains(t, "go test"),
		strings.Contains(t, "make health"),
		strings.Contains(t, "health check"),
		strings.Contains(t, "запустить тест"):
		return "health_check"
	default:
		// Any other todo — treated as todo_item; the worker will detect
		// "make <target>" patterns from the text and execute them.
		return "todo_item"
	}
}

