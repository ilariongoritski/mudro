package bot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTodoTasks(t *testing.T) {
	lines := []string{
		"- [ ] 2026-02-26 | P2 | area:bot | Добавить unit-тест",
		"  - контекст",
		"- [ ] 2026-02-26 | P2 | area:repo | gofmt",
	}
	tasks := parseTodoTasks(lines)
	if len(tasks) != 2 {
		t.Fatalf("len = %d, want 2", len(tasks))
	}
	if tasks[0].Complexity != 1 || tasks[0].Action != "go_test" {
		t.Fatalf("unexpected task0: %+v", tasks[0])
	}
	if tasks[1].Action != "gofmt_repo" {
		t.Fatalf("unexpected task1: %+v", tasks[1])
	}
}

func TestRewriteTodoWithout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todo.md")
	lines := []string{
		"- [ ] one",
		"",
		"- [ ] two",
	}
	done := []todoTask{{StartLine: 0, EndLine: 1}}
	if err := rewriteTodoWithout(path, lines, done); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.Contains(string(b), "one") {
		t.Fatalf("task not removed: %q", string(b))
	}
}

func TestAppendDone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "done.md")
	if err := os.WriteFile(path, []byte("# done\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := appendDone(path, []todoTask{{Title: "task x"}}); err != nil {
		t.Fatalf("appendDone: %v", err)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), "task x") {
		t.Fatalf("missing appended line: %q", string(b))
	}
}
