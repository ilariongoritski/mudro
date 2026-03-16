package bot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestLLMRuntimeTotals(t *testing.T) {
	rt := &runtimeTimeMemory{
		ByCommand: map[string]runtimeCommand{
			"mudro":      {Responses: 2, TotalMS: 3000, ProcessMS: 2500},
			"mudro_chat": {Responses: 1, TotalMS: 900, ProcessMS: 700},
			"health":     {Responses: 5, TotalMS: 1000, ProcessMS: 900},
		},
	}
	totalMS, processMS, responses := llmRuntimeTotals(rt)
	if totalMS != 3900 || processMS != 3200 || responses != 3 {
		t.Fatalf("unexpected llm totals: total=%d process=%d responses=%d", totalMS, processMS, responses)
	}
}

func TestRecordRuntimeTimePreservesBackfillSections(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}
	path := filepath.Join(repo, ".codex", "time_runtime.json")
	initial := `{
  "version": 1,
  "updated_at": "2026-03-09T22:53:07+03:00",
  "totals": {"responses": 1, "total_ms": 1000},
  "by_command": {"time": {"responses": 1, "total_ms": 1000}},
  "desktop_dialog_backfill": {"total_ms": 47308237, "turns": 225}
}`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("write initial runtime: %v", err)
	}

	if err := recordRuntimeTime(repo, "/health", 100*time.Millisecond, 50*time.Millisecond, 200*time.Millisecond); err != nil {
		t.Fatalf("recordRuntimeTime: %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated runtime: %v", err)
	}
	if !strings.Contains(string(updated), "desktop_dialog_backfill") {
		t.Fatalf("desktop backfill was removed: %s", string(updated))
	}

	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(updated, &raw); err != nil {
		t.Fatalf("unmarshal updated runtime: %v", err)
	}
	var backfill struct {
		TotalMS int64 `json:"total_ms"`
		Turns   int   `json:"turns"`
	}
	if err := json.Unmarshal(raw["desktop_dialog_backfill"], &backfill); err != nil {
		t.Fatalf("unmarshal backfill: %v", err)
	}
	if backfill.TotalMS != 47308237 || backfill.Turns != 225 {
		t.Fatalf("backfill changed unexpectedly: %+v", backfill)
	}
}
