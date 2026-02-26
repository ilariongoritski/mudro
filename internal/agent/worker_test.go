package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolation(t *testing.T) {
	if !IsUniqueViolation(&pgconn.PgError{Code: "23505"}) {
		t.Fatal("expected unique violation")
	}
	if IsUniqueViolation(errors.New("x")) {
		t.Fatal("unexpected unique violation")
	}
}

func TestProcessTodoTask(t *testing.T) {
	root := t.TempDir()
	w := &Worker{RepoRoot: root, WorkerID: "w1"}

	task := &Task{
		ID:      7,
		Kind:    "todo_item",
		Payload: []byte(`{"source":"todo.md","text":"do thing"}`),
	}
	if err := w.processTodoTask(task); err != nil {
		t.Fatalf("processTodoTask: %v", err)
	}

	logRoot := filepath.Join(root, ".codex", "logs")
	ents, err := os.ReadDir(logRoot)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(ents) == 0 {
		t.Fatal("expected created log dir")
	}
	b, err := os.ReadFile(filepath.Join(logRoot, ents[0].Name(), "agent-task.md"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(b), "Task ID: 7") {
		t.Fatalf("unexpected log content: %q", string(b))
	}
}

func TestProcessTaskUnsupported(t *testing.T) {
	w := &Worker{}
	err := w.processTask(context.Background(), &Task{Kind: "unknown"})
	if err == nil {
		t.Fatal("expected error for unsupported kind")
	}
}
