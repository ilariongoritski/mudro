package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (r *Runner) TodoList() ([]byte, error) {
	path := filepath.Join(r.RepoRoot, ".codex", "todo.md")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read todo: %w", err)
	}

	lines := strings.Split(string(b), "\n")
	items := make([]string, 0, 16)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [ ] ") {
			items = append(items, strings.TrimPrefix(line, "- [ ] "))
		}
	}

	if len(items) == 0 {
		return []byte("TODO пуст. Добавь пункт: /todoadd <текст>"), nil
	}

	var out strings.Builder
	out.WriteString("Будущие цели и улучшения:\n")
	for i, it := range items {
		out.WriteString(fmt.Sprintf("%d) %s\n", i+1, it))
		if i >= 11 {
			out.WriteString("...(сокращено)\n")
			break
		}
	}
	return []byte(strings.TrimSpace(out.String())), nil
}

func (r *Runner) TodoAdd(text string) ([]byte, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return []byte("Использование: /todoadd <цель или улучшение>"), nil
	}

	path := filepath.Join(r.RepoRoot, ".codex", "todo.md")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read todo: %w", err)
	}
	body := string(b)

	date := time.Now().Format("2006-01-02")
	entry := "- [ ] " + date + " | P2 | area:bot | " + text + "\n" +
		"  - Контекст: добавлено из Telegram `/todoadd`\n" +
		"  - Следующий шаг: декомпозировать и выполнить\n"

	if strings.Contains(body, entry) {
		return []byte("Такой пункт уже есть в TODO"), nil
	}

	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	body += entry
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return nil, fmt.Errorf("write todo: %w", err)
	}

	return []byte("Добавлено в TODO:\n- " + text), nil
}
