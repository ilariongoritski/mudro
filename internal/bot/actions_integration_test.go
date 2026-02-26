package bot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeRunLog(t *testing.T, root, runID, body string) {
	t.Helper()
	dir := filepath.Join(root, ".codex", "logs", runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir run dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write index.md: %v", err)
	}
}

func TestActionsSummaries(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".codex", "todo.md"), []byte("- [ ] 2026 | P2 | area:bot | тест\n"), 0o644); err != nil {
		t.Fatalf("write todo: %v", err)
	}
	now := time.Now().Add(-2 * time.Minute)
	runID := now.Format("20060102-1504")
	writeRunLog(t, root, runID, "Цель: Проверка\n- Ошибка: timeout\n- Решение: retry\n- Что прошло: успешно\n- Следующий шаг: сделать тесты\n")

	r := &Runner{RepoRoot: root}

	out10, err := r.ActionsLast10Min()
	if err != nil {
		t.Fatalf("ActionsLast10Min: %v", err)
	}
	if !strings.Contains(string(out10), "Проблемы:") {
		t.Fatalf("unexpected out10: %q", string(out10))
	}

	out1h, err := r.ActionsLast1H()
	if err != nil {
		t.Fatalf("ActionsLast1H: %v", err)
	}
	if !strings.Contains(string(out1h), "Путь за последний час:") {
		t.Fatalf("unexpected out1h: %q", string(out1h))
	}
}

func TestCollectHelpers(t *testing.T) {
	ins := []runInsight{
		{Problems: []string{"p1"}, Solutions: []string{"s1"}, Done: []string{"d1"}, Improvements: []string{"i1"}},
	}
	if len(collectProblems(ins)) != 1 || len(collectSolutions(ins)) != 1 || len(collectDone(ins)) != 1 || len(collectImprovements(ins)) != 1 {
		t.Fatalf("unexpected collect lengths")
	}
	if got := firstNonEmpty("  ", "x"); got != "x" {
		t.Fatalf("firstNonEmpty=%q", got)
	}
}
