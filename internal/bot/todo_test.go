package bot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTodoListAndAdd(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	todoPath := filepath.Join(root, ".codex", "todo.md")
	if err := os.WriteFile(todoPath, []byte("# todo\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	r := &Runner{RepoRoot: root}

	list, err := r.TodoList()
	if err != nil {
		t.Fatalf("TodoList empty: %v", err)
	}
	if !strings.Contains(string(list), "TODO пуст") {
		t.Fatalf("unexpected list: %q", string(list))
	}

	add, err := r.TodoAdd("новая задача")
	if err != nil {
		t.Fatalf("TodoAdd: %v", err)
	}
	if !strings.Contains(string(add), "Добавлено в TODO") {
		t.Fatalf("unexpected add output: %q", string(add))
	}

	list, err = r.TodoList()
	if err != nil {
		t.Fatalf("TodoList filled: %v", err)
	}
	if !strings.Contains(string(list), "новая задача") {
		t.Fatalf("missing task in list: %q", string(list))
	}
}
