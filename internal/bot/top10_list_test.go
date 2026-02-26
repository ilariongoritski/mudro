package bot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTop10List(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".codex", "top10.md"), []byte("1. one\nx\n2. two\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	r := &Runner{RepoRoot: root}
	out, err := r.Top10List()
	if err != nil {
		t.Fatalf("Top10List: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "1. one") || !strings.Contains(s, "2. two") {
		t.Fatalf("unexpected output: %q", s)
	}
}
