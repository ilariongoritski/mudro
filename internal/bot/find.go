package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type repoFinding struct {
	Priority int
	Title    string
	Details  string
	Todo     string
}

func (r *Runner) FindImprovements() ([]byte, error) {
	findings := make([]repoFinding, 0, 16)

	findings = append(findings, r.checkGoTestCoverageHint()...)
	findings = append(findings, r.findBackupArtifacts()...)
	findings = append(findings, r.findLargeGoFiles()...)
	findings = append(findings, r.findEmptyDirs()...)

	if len(findings) == 0 {
		return []byte("Улучшений не найдено"), nil
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Priority == findings[j].Priority {
			return findings[i].Title < findings[j].Title
		}
		return findings[i].Priority < findings[j].Priority
	})

	added := r.persistImportantFindings(findings)

	var out strings.Builder
	out.WriteString(fmt.Sprintf("Найдено улучшений: %d\n", len(findings)))
	for i, f := range findings {
		if i >= 12 {
			out.WriteString("...(сокращено)\n")
			break
		}
		out.WriteString(fmt.Sprintf("%d) [P%d] %s\n", i+1, f.Priority, f.Title))
		if f.Details != "" {
			out.WriteString("   " + f.Details + "\n")
		}
	}

	if len(added) > 0 {
		out.WriteString("Внесено в TODO (важнейшие):\n")
		for _, a := range added {
			out.WriteString("- " + a + "\n")
		}
	} else {
		out.WriteString("В TODO изменений не внесено (критичных новых пунктов нет).\n")
	}

	out.WriteString("Если хочешь, дам фокус-анализ отдельно по зонам: bot/api/db/docs.")
	return []byte(strings.TrimSpace(out.String())), nil
}

func (r *Runner) checkGoTestCoverageHint() []repoFinding {
	out, err := r.runStep([]string{"go", "test", "./..."})
	if err != nil {
		return []repoFinding{{
			Priority: 1,
			Title:    "Тесты не проходят",
			Details:  shortLine(string(out)),
			Todo:     "Починить падающие тесты `go test ./...`",
		}}
	}
	lines := strings.Split(string(out), "\n")
	noTests := 0
	for _, ln := range lines {
		if strings.Contains(ln, "[no test files]") {
			noTests++
		}
	}
	if noTests >= 5 {
		return []repoFinding{{
			Priority: 2,
			Title:    "Мало тестового покрытия",
			Details:  fmt.Sprintf("Пакетов без тестов: %d (по `go test ./...`)", noTests),
			Todo:     "Добавить unit-тесты на ключевые bot/api сценарии",
		}}
	}
	return nil
}

func (r *Runner) findBackupArtifacts() []repoFinding {
	root := r.RepoRoot
	patterns := []string{".bak", ".save", ".save.", "~"}
	var found []string

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		for _, p := range patterns {
			if strings.Contains(name, p) {
				found = append(found, rel)
				break
			}
		}
		return nil
	})

	if len(found) == 0 {
		return nil
	}
	sort.Strings(found)
	if len(found) > 6 {
		found = found[:6]
	}
	return []repoFinding{{
		Priority: 2,
		Title:    "Обнаружены артефакты backup/save в репозитории",
		Details:  strings.Join(found, ", "),
		Todo:     "Почистить или перенести backup/save артефакты из рабочей структуры",
	}}
}

func (r *Runner) findLargeGoFiles() []repoFinding {
	root := r.RepoRoot
	type item struct {
		path  string
		lines int
	}
	var large []item
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		n := 1 + strings.Count(string(b), "\n")
		if n >= 250 {
			large = append(large, item{path: rel, lines: n})
		}
		return nil
	})
	if len(large) == 0 {
		return nil
	}
	sort.Slice(large, func(i, j int) bool { return large[i].lines > large[j].lines })
	if len(large) > 3 {
		large = large[:3]
	}
	parts := make([]string, 0, len(large))
	for _, it := range large {
		parts = append(parts, fmt.Sprintf("%s (%d строк)", it.path, it.lines))
	}
	return []repoFinding{{
		Priority: 3,
		Title:    "Крупные Go-файлы требуют декомпозиции",
		Details:  strings.Join(parts, "; "),
		Todo:     "Декомпозировать крупные Go-файлы на модули по зонам ответственности",
	}}
}

func (r *Runner) findEmptyDirs() []repoFinding {
	root := r.RepoRoot
	var empty []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "." || strings.HasPrefix(rel, ".git") {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}
		if len(entries) == 0 {
			empty = append(empty, rel+"/")
		}
		return nil
	})
	if len(empty) == 0 {
		return nil
	}
	sort.Strings(empty)
	if len(empty) > 5 {
		empty = empty[:5]
	}
	return []repoFinding{{
		Priority: 4,
		Title:    "Есть пустые директории",
		Details:  strings.Join(empty, ", "),
		Todo:     "",
	}}
}

func (r *Runner) persistImportantFindings(findings []repoFinding) []string {
	path := filepath.Join(r.RepoRoot, ".codex", "todo.md")
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	text := string(body)
	added := make([]string, 0, 3)
	today := time.Now().Format("2006-01-02")

	for _, f := range findings {
		if f.Priority > 2 || f.Todo == "" {
			continue
		}
		if len(added) >= 3 {
			break
		}
		if strings.Contains(text, f.Todo) {
			continue
		}
		entry := fmt.Sprintf("- [ ] %s | P%d | area:repo | %s\n  - Контекст: авто-добавлено командой /find\n  - Следующий шаг: уточнить план и выполнить\n", today, f.Priority, f.Todo)
		if !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += entry
		added = append(added, f.Todo)
	}
	if len(added) > 0 {
		_ = os.WriteFile(path, []byte(text), 0o644)
	}
	return added
}
