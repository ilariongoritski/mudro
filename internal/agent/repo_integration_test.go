package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testAgentRepo(t *testing.T) (*Repository, *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, config.DSN())
	if err != nil {
		t.Skipf("skip integration test: db connect: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skip integration test: db ping: %v", err)
	}
	if _, err := pool.Exec(ctx, `create table if not exists agent_queue (
		id bigserial primary key,
		kind text not null,
		payload jsonb not null default '{}'::jsonb,
		status text not null default 'queued',
		priority int not null default 0,
		attempts int not null default 0,
		max_attempts int not null default 3,
		dedupe_key text,
		run_after timestamptz not null default now(),
		locked_by text,
		locked_at timestamptz,
		last_error text,
		created_at timestamptz not null default now(),
		updated_at timestamptz not null default now(),
		finished_at timestamptz
	)`); err != nil {
		t.Fatalf("ensure agent_queue: %v", err)
	}
	if _, err := pool.Exec(ctx, `create unique index if not exists agent_queue_dedupe_live_uq
		on agent_queue (dedupe_key)
		where dedupe_key is not null and status in ('queued','in_progress')`); err != nil {
		t.Fatalf("ensure dedupe index: %v", err)
	}
	if _, err := pool.Exec(ctx, `truncate table agent_queue restart identity`); err != nil {
		t.Fatalf("truncate agent_queue: %v", err)
	}
	return NewRepository(pool), pool
}

func TestRepositoryFlowIntegration(t *testing.T) {
	repo, pool := testAgentRepo(t)
	ctx := context.Background()

	if _, err := repo.Enqueue(ctx, "", map[string]any{}, 1, time.Time{}, 0, ""); err == nil {
		t.Fatal("expected error for empty kind")
	}

	id1, err := repo.Enqueue(ctx, "todo_item", map[string]any{"text": "A"}, 1, time.Now(), 3, "dup-1")
	if err != nil || id1 == 0 {
		t.Fatalf("enqueue id1: id=%d err=%v", id1, err)
	}
	id2, err := repo.Enqueue(ctx, "todo_item", map[string]any{"text": "B"}, 10, time.Now(), 3, "dup-2")
	if err != nil || id2 == 0 {
		t.Fatalf("enqueue id2: id=%d err=%v", id2, err)
	}
	dupID, err := repo.Enqueue(ctx, "todo_item", map[string]any{"text": "dup"}, 1, time.Now(), 3, "dup-1")
	if err != nil {
		t.Fatalf("enqueue dup: %v", err)
	}
	if dupID != 0 {
		t.Fatalf("expected duplicate enqueue id=0, got %d", dupID)
	}

	task, err := repo.ClaimNext(ctx, "w1")
	if err != nil || task == nil {
		t.Fatalf("claim next: task=%v err=%v", task, err)
	}
	if task.ID != id2 {
		t.Fatalf("expected high-priority task id=%d, got %d", id2, task.ID)
	}
	if err := repo.MarkDone(ctx, task.ID); err != nil {
		t.Fatalf("mark done: %v", err)
	}

	task2, err := repo.ClaimNext(ctx, "w1")
	if err != nil || task2 == nil {
		t.Fatalf("claim next #2: task=%v err=%v", task2, err)
	}
	if task2.ID != id1 {
		t.Fatalf("expected second task id=%d, got %d", id1, task2.ID)
	}
	if err := repo.MarkFailed(ctx, task2.ID, "boom", 0); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	var status, lastErr string
	if err := pool.QueryRow(ctx, `select status, coalesce(last_error,'') from agent_queue where id=$1`, task2.ID).Scan(&status, &lastErr); err != nil {
		t.Fatalf("select status: %v", err)
	}
	if status != StatusFailed || !strings.Contains(lastErr, "boom") {
		t.Fatalf("unexpected status=%q last_error=%q", status, lastErr)
	}
}

func TestPlanFromTodoIntegration(t *testing.T) {
	repo, pool := testAgentRepo(t)
	ctx := context.Background()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}
	body := strings.Join([]string{
		"# todo",
		"- [ ] 2026-02-26 | P2 | area:bot | first task",
		"- [x] completed task",
		"- [ ] second task",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(root, ".codex", "todo.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write todo.md: %v", err)
	}

	n, err := PlanFromTodo(ctx, root, repo)
	if err != nil {
		t.Fatalf("PlanFromTodo first: %v", err)
	}
	if n != 2 {
		t.Fatalf("enqueued first=%d, want 2", n)
	}

	n, err = PlanFromTodo(ctx, root, repo)
	if err != nil {
		t.Fatalf("PlanFromTodo second: %v", err)
	}
	if n != 0 {
		t.Fatalf("enqueued second=%d, want 0 due dedupe", n)
	}

	var cnt int
	if err := pool.QueryRow(ctx, `select count(*) from agent_queue`).Scan(&cnt); err != nil {
		t.Fatalf("count queue: %v", err)
	}
	if cnt != 2 {
		t.Fatalf("queue count=%d, want 2", cnt)
	}
}

func TestWorkerRunOnceIntegration(t *testing.T) {
	repo, pool := testAgentRepo(t)
	ctx := context.Background()

	w := &Worker{RepoRoot: t.TempDir(), Queue: repo, WorkerID: "test-worker"}

	processed, err := w.RunOnce(ctx)
	if err != nil {
		t.Fatalf("RunOnce empty queue err=%v", err)
	}
	if processed {
		t.Fatal("expected processed=false on empty queue")
	}

	_, err = repo.Enqueue(ctx, "todo_item", map[string]any{"source": "todo.md", "text": "do x"}, 1, time.Now(), 3, "worker-ok")
	if err != nil {
		t.Fatalf("enqueue ok task: %v", err)
	}
	processed, err = w.RunOnce(ctx)
	if err != nil || !processed {
		t.Fatalf("RunOnce ok processed=%v err=%v", processed, err)
	}

	_, err = repo.Enqueue(ctx, "todo_item", map[string]any{"source": "todo.md", "text": ""}, 1, time.Now(), 2, "worker-bad")
	if err != nil {
		t.Fatalf("enqueue bad task: %v", err)
	}
	processed, err = w.RunOnce(ctx)
	if !processed || err == nil {
		t.Fatalf("RunOnce bad expected error: processed=%v err=%v", processed, err)
	}

	var status string
	if err := pool.QueryRow(ctx, `select status from agent_queue where dedupe_key='worker-bad'`).Scan(&status); err != nil {
		t.Fatalf("select worker-bad status: %v", err)
	}
	if status != StatusQueued {
		t.Fatalf("worker-bad status=%q, want queued for retry", status)
	}
}
