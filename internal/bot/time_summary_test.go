package bot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRebuildMemoryJSONAndTimeSummary(t *testing.T) {
	root := t.TempDir()
	logs := filepath.Join(root, ".codex", "logs")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatalf("mkdir logs: %v", err)
	}

	now := time.Now()
	for _, d := range []time.Time{now.Add(-15 * time.Minute), now.Add(-5 * time.Minute)} {
		if err := os.MkdirAll(filepath.Join(logs, d.Format("20060102-1504")), 0o755); err != nil {
			t.Fatalf("mkdir run: %v", err)
		}
	}

	if err := os.WriteFile(filepath.Join(root, ".codex", "time_runtime.json"), []byte(`{
  "version": 1,
  "totals": {"responses": 2, "total_ms": 1200, "process_ms": 1000, "send_ms": 100, "unknown_ms": 100, "max_total_ms": 700, "last_total_ms": 500},
  "by_command": {"health": {"responses": 2, "total_ms": 1200, "process_ms": 1000, "send_ms": 100, "unknown_ms": 100, "max_total_ms": 700, "last_total_ms": 500}}
}`), 0o644); err != nil {
		t.Fatalf("write runtime: %v", err)
	}

	r := &Runner{RepoRoot: root}
	mem, err := r.rebuildMemoryJSON()
	if err != nil {
		t.Fatalf("rebuildMemoryJSON: %v", err)
	}
	if mem.Totals.Runs == 0 {
		t.Fatalf("expected runs > 0")
	}

	out, err := r.TimeSummary()
	if err != nil {
		t.Fatalf("TimeSummary: %v", err)
	}
	if !strings.Contains(string(out), "Время работы:") {
		t.Fatalf("unexpected summary: %q", string(out))
	}
}
