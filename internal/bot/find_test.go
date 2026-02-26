package bot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindHelpersAndPersist(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".codex", "todo.md"), []byte("# todo\n"), 0o644); err != nil {
		t.Fatalf("write todo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "x.save"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write save: %v", err)
	}
	large := strings.Repeat("package p\n", 260)
	if err := os.WriteFile(filepath.Join(root, "big.go"), []byte(large), 0o644); err != nil {
		t.Fatalf("write big.go: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatalf("mkdir empty: %v", err)
	}

	r := &Runner{RepoRoot: root}
	if len(r.findBackupArtifacts()) == 0 {
		t.Fatal("expected backup artifact finding")
	}
	if len(r.findLargeGoFiles()) == 0 {
		t.Fatal("expected large file finding")
	}
	if len(r.findEmptyDirs()) == 0 {
		t.Fatal("expected empty dir finding")
	}

	added := r.persistImportantFindings([]repoFinding{{
		Priority: 2,
		Todo:     "Добавить unit-тесты на ключевые bot/api сценарии",
	}})
	if len(added) != 1 {
		t.Fatalf("added=%v", added)
	}
}

func TestFindImprovements(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".codex", "todo.md"), []byte("# todo\n"), 0o644); err != nil {
		t.Fatalf("write todo: %v", err)
	}
	// Trigger at least one deterministic finding.
	if err := os.WriteFile(filepath.Join(root, "x.save"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write save: %v", err)
	}

	r := &Runner{RepoRoot: root}
	out, err := r.FindImprovements()
	if err != nil {
		t.Fatalf("FindImprovements: %v", err)
	}
	if !strings.Contains(string(out), "Найдено улучшений") {
		t.Fatalf("unexpected output: %q", string(out))
	}
}
