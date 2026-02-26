package bot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetChatModeStore() {
	chatModeStore.mu.Lock()
	defer chatModeStore.mu.Unlock()
	chatModeStore.ready = false
	chatModeStore.data = chatModeFile{Chats: map[string]bool{}}
}

func TestChatModeSetAndStatus(t *testing.T) {
	resetChatModeStore()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	r := &Runner{RepoRoot: root}

	out, err := r.SetChatMode(123, true)
	if err != nil {
		t.Fatalf("SetChatMode: %v", err)
	}
	if !strings.Contains(string(out), "включен") {
		t.Fatalf("unexpected set output: %q", string(out))
	}
	if !r.isChatModeEnabled(123) {
		t.Fatal("mode should be enabled")
	}

	t.Setenv("OPENAI_API_KEY", "")
	st, err := r.ChatModeStatus(123)
	if err != nil {
		t.Fatalf("ChatModeStatus: %v", err)
	}
	if !strings.Contains(string(st), "OPENAI_API_KEY не задан") {
		t.Fatalf("status output: %q", string(st))
	}

	if _, err := r.SetChatMode(123, false); err != nil {
		t.Fatalf("disable mode: %v", err)
	}
	st, err = r.ChatModeStatus(123)
	if err != nil {
		t.Fatalf("ChatModeStatus off: %v", err)
	}
	if !strings.Contains(string(st), "OFF") {
		t.Fatalf("status output: %q", string(st))
	}
}

func TestHandleChatText(t *testing.T) {
	r := &Runner{}
	if _, ok := r.handleChatText("   "); ok {
		t.Fatal("empty text should be ignored")
	}
	if _, ok := r.handleChatText("/help"); ok {
		t.Fatal("command should be ignored")
	}
	if txt, ok := r.handleChatText("привет"); !ok || txt != "привет" {
		t.Fatalf("unexpected: txt=%q ok=%v", txt, ok)
	}
}
