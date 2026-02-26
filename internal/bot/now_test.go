package bot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadDoneItems(t *testing.T) {
	p := filepath.Join(t.TempDir(), "done.md")
	body := "x\n- 2026-02-25 | one\n- YYYY-MM-DD | template\n- 2026-02-26 | two\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := readDoneItems(p, 5)
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
}

func TestAppendToTodoAndShortLine(t *testing.T) {
	p := filepath.Join(t.TempDir(), "todo.md")
	if err := os.WriteFile(p, []byte("head\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	item := "- [ ] item\n"
	if !appendToTodo(p, item) {
		t.Fatal("appendToTodo first should append")
	}
	if appendToTodo(p, item) {
		t.Fatal("appendToTodo duplicate should be false")
	}

	if got := shortLine("a\nb"); got != "a" {
		t.Fatalf("shortLine=%q", got)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "item") {
		t.Fatalf("todo body: %q", string(b))
	}
}
