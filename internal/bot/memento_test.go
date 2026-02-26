package bot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffFilesAndTrimList(t *testing.T) {
	added, removed := diffFiles([]string{"a", "b"}, []string{"b", "c"})
	if len(added) != 1 || added[0] != "c" || len(removed) != 1 || removed[0] != "a" {
		t.Fatalf("added=%v removed=%v", added, removed)
	}
	if got := trimList([]string{"1", "2", "3"}, 2); len(got) != 2 {
		t.Fatalf("trimList len=%d", len(got))
	}
}

func TestMementoHelpers(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}
	r := &Runner{RepoRoot: root}

	created, err := r.ensureMemoryFiles()
	if err != nil {
		t.Fatalf("ensureMemoryFiles: %v", err)
	}
	if len(created) == 0 {
		t.Fatal("expected memory files created")
	}

	// Required for checksums.
	if err := os.WriteFile(filepath.Join(root, ".codex", "memory.json"), []byte(`{"version":1}`), 0o644); err != nil {
		t.Fatalf("write memory.json: %v", err)
	}
	if _, err := r.memoryChecksums(); err != nil {
		t.Fatalf("memoryChecksums: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "one.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	files, err := r.collectRepoFiles()
	if err != nil {
		t.Fatalf("collectRepoFiles: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected repo files")
	}

	if err := r.writeMemento(mementoJSON{Version: 1, RepoFiles: files}); err != nil {
		t.Fatalf("writeMemento: %v", err)
	}
	prev, err := r.readPrevMemento()
	if err != nil || prev.Version != 1 {
		t.Fatalf("readPrevMemento err=%v prev=%+v", err, prev)
	}

	addedTodos, err := r.addMementoTodoIfNeeded([]string{"new.go"}, nil)
	if err != nil {
		t.Fatalf("addMementoTodoIfNeeded: %v", err)
	}
	if len(addedTodos) != 1 {
		t.Fatalf("addedTodos=%v", addedTodos)
	}
	b, _ := os.ReadFile(filepath.Join(root, ".codex", "todo.md"))
	if !strings.Contains(string(b), "Проверить и классифицировать новые файлы") {
		t.Fatalf("todo not updated: %q", string(b))
	}
}
