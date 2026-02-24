package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (r *Runner) RepoStructure() ([]byte, error) {
	const (
		maxDepth = 3
		maxLines = 180
	)

	ignoreDirs := map[string]struct{}{
		".git":   {},
		".bin":   {},
		"tmp":    {},
		"out":    {},
		"var":    {},
		"vendor": {},
	}

	lines := []string{"Структура репозитория:"}
	var walk func(dir string, depth int, prefix string) error
	walk = func(dir string, depth int, prefix string) error {
		if depth > maxDepth || len(lines) >= maxLines {
			return nil
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		sort.Slice(entries, func(i, j int) bool {
			a, b := entries[i], entries[j]
			if a.IsDir() && !b.IsDir() {
				return true
			}
			if !a.IsDir() && b.IsDir() {
				return false
			}
			return a.Name() < b.Name()
		})

		for _, e := range entries {
			name := e.Name()
			if depth == 0 && name == ".codex" {
				lines = append(lines, prefix+name+"/")
				lines = append(lines, prefix+"  done.md")
				lines = append(lines, prefix+"  todo.md")
				lines = append(lines, prefix+"  top10.md")
				continue
			}
			if e.IsDir() {
				if _, skip := ignoreDirs[name]; skip {
					continue
				}
				lines = append(lines, prefix+name+"/")
				if len(lines) >= maxLines {
					return nil
				}
				if err := walk(filepath.Join(dir, name), depth+1, prefix+"  "); err != nil {
					return err
				}
				continue
			}
			lines = append(lines, prefix+name)
			if len(lines) >= maxLines {
				return nil
			}
		}
		return nil
	}

	if err := walk(r.RepoRoot, 0, ""); err != nil {
		return nil, fmt.Errorf("scan repo: %w", err)
	}
	if len(lines) >= maxLines {
		lines = append(lines, "...(сокращено)")
	}
	return []byte(strings.Join(lines, "\n")), nil
}
