package bot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
)

func (r *Runner) ActionsLast10Min() ([]byte, error) {
	runs, err := r.selectRunsSince(10 * time.Minute)
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return []byte("за последние 10 минут записей нет"), nil
	}

	insights := make([]runInsight, 0, len(runs))
	for _, run := range runs {
		ri, err := parseRunInsight(run.ts, run.dir)
		if err != nil {
			continue
		}
		insights = append(insights, ri)
	}
	if len(insights) == 0 {
		return []byte("не удалось разобрать логи за последние 10 минут"), nil
	}

	var b strings.Builder
	b.WriteString("Задача:\n")
	b.WriteString("- " + firstNonEmpty(insights[len(insights)-1].Goal, "операционная работа по проекту") + "\n")

	b.WriteString("Проблемы:\n")
	problems := uniqueNonEmpty(collectProblems(insights), 4)
	if len(problems) == 0 {
		b.WriteString("- критичных проблем не зафиксировано\n")
	} else {
		for _, p := range problems {
			b.WriteString("- " + p + "\n")
		}
	}

	b.WriteString("Выполнения:\n")
	done := uniqueNonEmpty(collectDone(insights), 5)
	if len(done) == 0 {
		b.WriteString("- зафиксированы действия, но без явного списка результатов\n")
	} else {
		for _, d := range done {
			b.WriteString("- " + d + "\n")
		}
	}

	b.WriteString("Возможные доработки:\n")
	improvements := uniqueNonEmpty(collectImprovements(insights), 4)
	if len(improvements) == 0 {
		improvements = readTodoHints(filepath.Join(r.RepoRoot, ".codex", "todo.md"), 4)
	}
	if len(improvements) == 0 {
		b.WriteString("- добавить явный следующий шаг в каждый прогон\n")
	} else {
		for _, item := range improvements {
			b.WriteString("- " + item + "\n")
		}
	}

	return []byte(strings.TrimSpace(b.String())), nil
}

func (r *Runner) ActionsLast1H() ([]byte, error) {
	runs, err := r.selectRunsSince(1 * time.Hour)
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return []byte("за последний час записей нет"), nil
	}

	insights := make([]runInsight, 0, len(runs))
	for _, run := range runs {
		ri, err := parseRunInsight(run.ts, run.dir)
		if err != nil {
			continue
		}
		insights = append(insights, ri)
	}
	if len(insights) == 0 {
		return []byte("не удалось разобрать логи за последний час"), nil
	}

	var b strings.Builder
	b.WriteString("Путь за последний час:\n")
	for _, ri := range insights {
		b.WriteString(fmt.Sprintf("- %s: %s\n", ri.TS.Format("15:04"), firstNonEmpty(ri.Goal, "рабочий прогон")))
		if len(ri.Problems) > 0 {
			b.WriteString("  Проблема: " + ri.Problems[0] + "\n")
		}
		if len(ri.Solutions) > 0 {
			b.WriteString("  Решение: " + ri.Solutions[0] + "\n")
		}
		if len(ri.Done) > 0 {
			b.WriteString("  Результат: " + ri.Done[0] + "\n")
		}
	}

	aggProblems := uniqueNonEmpty(collectProblems(insights), 3)
	aggSolutions := uniqueNonEmpty(collectSolutions(insights), 3)
	if len(aggProblems) > 0 {
		b.WriteString("Итоговые проблемы:\n")
		for _, p := range aggProblems {
			b.WriteString("- " + p + "\n")
		}
	}
	if len(aggSolutions) > 0 {
		b.WriteString("Итоговые решения:\n")
		for _, s := range aggSolutions {
			b.WriteString("- " + s + "\n")
		}
	}

	return []byte(strings.TrimSpace(b.String())), nil
}

type runDir struct {
	ts  time.Time
	dir string
}

func (r *Runner) selectRunsSince(d time.Duration) ([]runDir, error) {
	logDir := filepath.Join(r.RepoRoot, config.CodexLogsDir())
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("read logs dir: %w", err)
	}
	cutoff := time.Now().Add(-d)

	runs := make([]runDir, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ts, err := time.Parse("20060102-1504", e.Name())
		if err != nil {
			continue
		}
		if ts.Before(cutoff) {
			continue
		}
		runs = append(runs, runDir{ts: ts, dir: filepath.Join(logDir, e.Name())})
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].ts.Before(runs[j].ts) })
	return runs, nil
}

type runInsight struct {
	TS           time.Time
	Goal         string
	Problems     []string
	Solutions    []string
	Done         []string
	Improvements []string
}

func parseRunInsight(ts time.Time, runDir string) (runInsight, error) {
	path := filepath.Join(runDir, "index.md")
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return runInsight{}, fmt.Errorf("index.md not found in %s", runDir)
		}
		return runInsight{}, err
	}

	lines := strings.Split(string(b), "\n")
	ri := runInsight{TS: ts}
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)

		if strings.Contains(lower, "цель:") && ri.Goal == "" {
			ri.Goal = cleanLine(afterColon(line))
			continue
		}
		if strings.HasPrefix(line, "- ") && ri.Goal == "" && strings.Contains(lower, "задача") {
			ri.Goal = cleanLine(strings.TrimPrefix(line, "- "))
		}

		if isProblemLine(lower) {
			ri.Problems = append(ri.Problems, cleanLine(stripBullet(line)))
			continue
		}
		if isSolutionLine(lower) {
			ri.Solutions = append(ri.Solutions, cleanLine(stripBullet(line)))
			continue
		}
		if isDoneLine(lower) {
			ri.Done = append(ri.Done, cleanLine(stripBullet(line)))
			continue
		}
		if strings.Contains(lower, "следующий шаг:") || strings.Contains(lower, "возможные доработки") {
			ri.Improvements = append(ri.Improvements, cleanLine(afterColon(line)))
		}
	}

	ri.Problems = uniqueNonEmpty(ri.Problems, 3)
	ri.Solutions = uniqueNonEmpty(ri.Solutions, 3)
	ri.Done = uniqueNonEmpty(ri.Done, 3)
	ri.Improvements = uniqueNonEmpty(ri.Improvements, 3)
	return ri, nil
}

func isProblemLine(lower string) bool {
	if strings.Contains(lower, "ошибка") || strings.Contains(lower, "error") || strings.Contains(lower, "failed") || strings.Contains(lower, "упало") || strings.Contains(lower, "проблем") {
		return !strings.Contains(lower, "нет")
	}
	return false
}

func isSolutionLine(lower string) bool {
	return strings.Contains(lower, "решени") || strings.Contains(lower, "починил") || strings.Contains(lower, "fix")
}

func isDoneLine(lower string) bool {
	return strings.Contains(lower, "успешно") || strings.Contains(lower, "что прошло") || strings.Contains(lower, "выполн") || strings.Contains(lower, "добавлен") || strings.Contains(lower, "изменен")
}

func collectProblems(in []runInsight) []string {
	out := make([]string, 0, 8)
	for _, ri := range in {
		out = append(out, ri.Problems...)
	}
	return out
}

func collectSolutions(in []runInsight) []string {
	out := make([]string, 0, 8)
	for _, ri := range in {
		out = append(out, ri.Solutions...)
	}
	return out
}

func collectDone(in []runInsight) []string {
	out := make([]string, 0, 8)
	for _, ri := range in {
		out = append(out, ri.Done...)
	}
	return out
}

func collectImprovements(in []runInsight) []string {
	out := make([]string, 0, 8)
	for _, ri := range in {
		out = append(out, ri.Improvements...)
	}
	return out
}

func readTodoHints(path string, max int) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(b), "\n")
	out := make([]string, 0, max)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "- [ ] ") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) == 0 {
			continue
		}
		item := strings.TrimSpace(parts[len(parts)-1])
		item = strings.TrimPrefix(item, "- [ ] ")
		item = cleanLine(item)
		if item == "" {
			continue
		}
		out = append(out, item)
		if len(out) >= max {
			break
		}
	}
	return out
}

func uniqueNonEmpty(items []string, max int) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, it := range items {
		it = cleanLine(it)
		if it == "" {
			continue
		}
		if _, ok := seen[it]; ok {
			continue
		}
		seen[it] = struct{}{}
		out = append(out, it)
		if max > 0 && len(out) >= max {
			break
		}
	}
	return out
}

func firstNonEmpty(v string, fallback string) string {
	v = cleanLine(v)
	if v == "" {
		return fallback
	}
	return v
}

func afterColon(s string) string {
	i := strings.Index(s, ":")
	if i < 0 || i+1 >= len(s) {
		return s
	}
	return strings.TrimSpace(s[i+1:])
}

func stripBullet(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "- ")
	return strings.TrimSpace(s)
}

func cleanLine(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "`")
	s = strings.TrimSpace(s)
	return s
}
