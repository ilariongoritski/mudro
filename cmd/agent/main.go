package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/internal/agent"
	"github.com/goritskimihail/mudro/internal/config"
)

func main() {
	mode := flag.String("mode", "planner", "planner|worker|once|approve|reject")
	interval := flag.Duration("interval", 1*time.Minute, "loop interval")
	workerID := flag.String("worker-id", "agent-worker-1", "worker id for queue locks")
	taskID := flag.Int64("task-id", 0, "task id for approve/reject mode")
	reason := flag.String("reason", "", "reject reason")
	flag.Parse()

	repoRoot := config.RepoRoot()
	dsn := config.DSN()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	q := agent.NewRepository(pool)
	w := &agent.Worker{RepoRoot: repoRoot, Queue: q, WorkerID: *workerID}

	switch *mode {
	case "planner":
		runPlannerLoop(repoRoot, q, *interval)
	case "worker":
		runWorkerLoop(w, *interval)
	case "once":
		runPlannerOnce(repoRoot, q)
		if _, err := w.RunOnce(context.Background()); err != nil {
			log.Printf("worker run once: %v", err)
		}
	case "approve":
		if *taskID <= 0 {
			log.Fatal("approve mode requires --task-id > 0")
		}
		if err := q.ApproveTask(context.Background(), *taskID); err != nil {
			log.Fatalf("approve task: %v", err)
		}
		log.Printf("approved task id=%d", *taskID)
	case "reject":
		if *taskID <= 0 {
			log.Fatal("reject mode requires --task-id > 0")
		}
		if err := q.RejectTask(context.Background(), *taskID, *reason); err != nil {
			log.Fatalf("reject task: %v", err)
		}
		log.Printf("rejected task id=%d", *taskID)
	default:
		log.Fatalf("unsupported mode: %s", *mode)
	}
}

func runPlannerLoop(repoRoot string, q *agent.Repository, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}

	runPlannerOnce(repoRoot, q)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		runPlannerOnce(repoRoot, q)
	}
}

func runPlannerOnce(repoRoot string, q *agent.Repository) {
	n, err := agent.PlanFromTodo(context.Background(), repoRoot, q)
	if err != nil {
		log.Printf("planner error: %v", err)
		return
	}
	log.Printf("planner: enqueued %d tasks", n)
}

func runWorkerLoop(w *agent.Worker, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}

	for {
		processed, err := w.RunOnce(context.Background())
		if err != nil {
			log.Printf("worker error: %v", err)
		}
		if !processed {
			time.Sleep(interval)
		}
	}
}
