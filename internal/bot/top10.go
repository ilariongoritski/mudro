package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r *Runner) Top10List() ([]byte, error) {
	path := filepath.Join(r.RepoRoot, ".codex", "top10.md")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read top10: %w", err)
	}

	lines := strings.Split(string(b), "\n")
	items := make([]string, 0, 10)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if isTop10Line(line) {
			items = append(items, line)
		}
	}
	if len(items) == 0 {
		return []byte("TOP10 пока пуст"), nil
	}

	var out strings.Builder
	out.WriteString("TOP10 значимых изменений:\n")
	for _, it := range items {
		out.WriteString("- " + it + "\n")
	}
	return []byte(strings.TrimSpace(out.String())), nil
}

func isTop10Line(line string) bool {
	if len(line) < 3 {
		return false
	}
	if line[0] < '1' || line[0] > '9' {
		return false
	}
	if strings.HasPrefix(line, "10.") {
		return true
	}
	return len(line) > 1 && line[1] == '.'
}
