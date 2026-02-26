package bot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTGLogHelpers(t *testing.T) {
	if got := nonEmpty("", "x"); got != "x" {
		t.Fatalf("nonEmpty=%q", got)
	}
	if got := trimForLog("  abc  ", 3); got != "abc" {
		t.Fatalf("trimForLog=%q", got)
	}
	if got := shortTS("bad-ts"); got != "bad-ts" {
		t.Fatalf("shortTS=%q", got)
	}
}

func TestAppendTGControlEventAndRead(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := appendTGControlEvent(repo, tgControlEvent{
		Username: "u",
		ChatID:   1,
		Command:  "/health",
		Status:   "ok",
	}); err != nil {
		t.Fatalf("appendTGControlEvent: %v", err)
	}

	r := &Runner{RepoRoot: repo}
	out, err := r.TelegramControlLog()
	if err != nil {
		t.Fatalf("TelegramControlLog: %v", err)
	}
	if !strings.Contains(string(out), "/health") {
		t.Fatalf("unexpected output: %q", string(out))
	}
}
