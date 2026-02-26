package reporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpers(t *testing.T) {
	if got := toInt64(float64(12)); got != 12 {
		t.Fatalf("toInt64=%d", got)
	}
	if got := fmtDuration(3661); got != "01:01:01" {
		t.Fatalf("fmtDuration=%q", got)
	}
	if got := trimAfter("abcdef", 3); got != "abc..." {
		t.Fatalf("trimAfter=%q", got)
	}
}

func TestParseMemoryAndRuntime(t *testing.T) {
	mem := map[string]any{
		"totals": map[string]any{"total_seconds": float64(600), "runs": float64(3)},
		"days":   map[string]any{},
	}
	_, total, runs := parseMemory(mem)
	if total != 600 || runs != 3 {
		t.Fatalf("parseMemory total=%d runs=%d", total, runs)
	}

	rt := map[string]any{"totals": map[string]any{"responses": float64(2), "total_ms": float64(1200)}}
	resp, ms := parseRuntime(rt)
	if resp != 2 || ms != 1200 {
		t.Fatalf("parseRuntime resp=%d ms=%d", resp, ms)
	}
}

func TestBuildSummaryAndResolveChatID(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex", "logs", "20260226-0000"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	write := func(rel string, body string) {
		t.Helper()
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	write(".codex/top10.md", "1. item one\n2. item two\n")
	write(".codex/todo.md", "- [ ] todo one\n")
	write(".codex/memory.json", `{"totals":{"total_seconds":1200,"runs":4},"days":{}}`)
	write(".codex/time_runtime.json", `{"totals":{"responses":2,"total_ms":4000}}`)

	ev := map[string]any{"chat_id": float64(123)}
	b, _ := json.Marshal(ev)
	write(".codex/tg_control.jsonl", string(b)+"\n")

	sum, err := BuildSummary(root)
	if err != nil {
		t.Fatalf("BuildSummary: %v", err)
	}
	if !strings.Contains(sum.Text, "Mudro reporter") {
		t.Fatalf("summary text: %q", sum.Text)
	}
	if id := ResolveChatID(root, 0); id != 123 {
		t.Fatalf("ResolveChatID=%d", id)
	}
}
