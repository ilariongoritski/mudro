package bot

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUpdateKeyAndNonNegativeMS(t *testing.T) {
	if got := updateKey(12, 34); got != "12:34" {
		t.Fatalf("updateKey=%q", got)
	}
	if got := nonNegativeMS(-time.Second); got != 0 {
		t.Fatalf("nonNegativeMS=%d", got)
	}
}

func TestRecordRuntimeTime(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}
	if err := recordRuntimeTime(repo, "/health", 120*time.Millisecond, 30*time.Millisecond, 200*time.Millisecond); err != nil {
		t.Fatalf("recordRuntimeTime: %v", err)
	}
	mem, err := readRuntimeTime(filepath.Join(repo, ".codex", "time_runtime.json"))
	if err != nil {
		t.Fatalf("readRuntimeTime: %v", err)
	}
	if mem.Totals.Responses != 1 || mem.ByCommand["health"].Responses != 1 {
		t.Fatalf("bad counters: %+v", mem)
	}
}
