package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
)

func (r *Runner) NowSummary() ([]byte, error) {
	firstRun, runsCount := r.detectRunWindow()
	since := firstRun
	if since.IsZero() {
		since = time.Now().AddDate(0, 0, -7)
	}

	commits, filesChanged, insertions, deletions, err := r.gitStatsSince(since)
	if err != nil {
		return nil, err
	}
	doneItems := readDoneItems(filepath.Join(r.RepoRoot, ".codex", "done.md"), 5)
	improvements := r.saveSuggestedImprovements()

	_, dbcheckErr := r.runStep([]string{"make", "dbcheck"})
	stateNow := "ok"
	if dbcheckErr != nil {
		stateNow = "degraded"
	}

	var out strings.Builder
	out.WriteString("Итог с запуска:\n")
	if !firstRun.IsZero() {
		out.WriteString(fmt.Sprintf("- Старт отсчета: %s\n", firstRun.Format("2006-01-02 15:04")))
	}
	out.WriteString(fmt.Sprintf("- Прогонов зафиксировано: %d\n", runsCount))
	out.WriteString(fmt.Sprintf("- Коммитов: %d | Файлов изменено: %d | Строк: +%d/-%d\n", commits, filesChanged, insertions, deletions))

	out.WriteString("Сделано:\n")
	if len(doneItems) == 0 {
		out.WriteString("- История выполненного пока не заполнена\n")
	} else {
		for _, item := range doneItems {
			out.WriteString("- " + item + "\n")
		}
	}

	out.WriteString("Состояние сейчас:\n")
	out.WriteString("- БД check: " + stateNow + "\n")
	if dbcheckErr != nil {
		out.WriteString("- Причина: " + shortLine(dbcheckErr.Error()) + "\n")
	}

	if len(improvements) > 0 {
		out.WriteString("Добавлено в TODO:\n")
		for _, imp := range improvements {
			out.WriteString("- " + imp + "\n")
		}
	}

	return []byte(strings.TrimSpace(out.String())), nil
}

func (r *Runner) detectRunWindow() (time.Time, int) {
	logDir := filepath.Join(r.RepoRoot, config.CodexLogsDir())
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return time.Time{}, 0
	}
	runs := make([]time.Time, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ts, err := time.Parse("20060102-1504", e.Name())
		if err != nil {
			continue
		}
		runs = append(runs, ts)
	}
	if len(runs) == 0 {
		return time.Time{}, 0
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].Before(runs[j]) })
	return runs[0], len(runs)
}

func (r *Runner) gitStatsSince(since time.Time) (commits int, filesChanged int, insertions int, deletions int, err error) {
	sinceArg := "--since=" + since.Format("2006-01-02 15:04:05")

	logOut, err := r.runStep([]string{"git", "log", sinceArg, "--pretty=format:%h"})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	for _, line := range strings.Split(string(logOut), "\n") {
		if strings.TrimSpace(line) != "" {
			commits++
		}
	}

	shortstatOut, err := r.runStep([]string{"git", "log", sinceArg, "--shortstat", "--pretty=format:"})
	if err != nil {
		return commits, 0, 0, 0, err
	}
	filesChanged, insertions, deletions = parseShortstat(string(shortstatOut))
	return commits, filesChanged, insertions, deletions, nil
}

func readDoneItems(path string, max int) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(b), "\n")
	items := make([]string, 0, max)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		if strings.HasPrefix(strings.ToLower(line), "yyyy-") || strings.HasPrefix(strings.ToLower(line), "выполнено:") {
			continue
		}
		items = append(items, line)
		if len(items) >= max {
			break
		}
	}
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return items
}

func (r *Runner) saveSuggestedImprovements() []string {
	suggestions := make([]string, 0, 2)
	todoPath := filepath.Join(r.RepoRoot, ".codex", "todo.md")
	body, _ := os.ReadFile(todoPath)
	todoText := strings.ToLower(string(body))
	today := time.Now().Format("2006-01-02")

	if !strings.Contains(todoText, "area:bot") || !strings.Contains(todoText, "тест") {
		item := "- [ ] " + today + " | P2 | area:bot | Добавить unit-тесты для /actions10, /actions1h, /commits3\n" +
			"  - Контекст: сводки строятся эвристиками, нужен контроль регрессий\n" +
			"  - Следующий шаг: покрыть парсинг run-логов и коммит-сводок таблицей кейсов\n"
		if appendToTodo(todoPath, item) {
			suggestions = append(suggestions, "unit-тесты для новых сводок бота")
		}
	}

	return suggestions
}

func appendToTodo(path string, item string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	text := string(b)
	if strings.Contains(text, item) {
		return false
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	text += item
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return false
	}
	return true
}

func shortLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return s
}
