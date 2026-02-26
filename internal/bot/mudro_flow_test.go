package bot

import (
	"testing"
	"time"
)

func TestMudroFlowChoiceStore(t *testing.T) {
	chatID := int64(42)
	clearPendingChoice(chatID)

	savePendingChoice(chatID, " q ", " a ")
	got, ok := loadPendingChoice(chatID)
	if !ok || got.question != "q" || got.answer != "a" {
		t.Fatalf("unexpected pending choice: ok=%v v=%+v", ok, got)
	}

	clearPendingChoice(chatID)
	if _, ok := loadPendingChoice(chatID); ok {
		t.Fatal("expected no pending choice after clear")
	}
}

func TestMudroFlowExpiry(t *testing.T) {
	chatID := int64(77)
	flowStore.mu.Lock()
	flowStore.byChat[chatID] = pendingChoice{
		question: "q",
		answer:   "a",
		created:  time.Now().Add(-21 * time.Minute),
	}
	flowStore.mu.Unlock()

	if _, ok := loadPendingChoice(chatID); ok {
		t.Fatal("expected expired choice to be dropped")
	}
}

func TestChoiceParsing(t *testing.T) {
	if !detectChoicePrompt("1. A\n2. B") {
		t.Fatal("expected detected prompt")
	}
	if detectChoicePrompt("A only") {
		t.Fatal("unexpected prompt detection")
	}

	if c, ok := isChoiceReply(" 1 "); !ok || c != "1" {
		t.Fatalf("isChoiceReply got=%q ok=%v", c, ok)
	}
	if _, ok := isChoiceReply("3"); ok {
		t.Fatal("expected invalid choice")
	}
}
