package bot

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseRunInsight(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte(`
Цель: Проверка API
- Ошибка: timeout
- Решение: повторил make up
- Что прошло: миграции успешно
- Следующий шаг: добавить тесты
`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	ri, err := parseRunInsight(time.Now(), dir)
	if err != nil {
		t.Fatalf("parseRunInsight: %v", err)
	}
	if ri.Goal != "Проверка API" {
		t.Fatalf("goal = %q", ri.Goal)
	}
	if len(ri.Problems) == 0 || len(ri.Solutions) == 0 || len(ri.Done) == 0 || len(ri.Improvements) == 0 {
		t.Fatalf("unexpected insight: %+v", ri)
	}
}

func TestUniqueNonEmpty(t *testing.T) {
	got := uniqueNonEmpty([]string{" a ", "a", "", "b"}, 10)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got = %#v", got)
	}
}

func TestReadTodoHints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todo.md")
	body := "- [ ] 2026-02-26 | P2 | area:bot | Добавить тесты\n- [x] done\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := readTodoHints(path, 3)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}
