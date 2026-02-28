package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Enqueue(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
	return r.enqueueWithStatus(ctx, kind, payload, priority, runAfter, maxAttempts, dedupeKey, StatusQueued)
}

func (r *Repository) EnqueueWaitingApproval(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
	return r.enqueueWithStatus(ctx, kind, payload, priority, runAfter, maxAttempts, dedupeKey, StatusWaitingApproval)
}

func (r *Repository) enqueueWithStatus(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey, status string) (int64, error) {
	if kind == "" {
		return 0, errors.New("kind is required")
	}
	if status == "" {
		status = StatusQueued
	}
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if runAfter.IsZero() {
		runAfter = time.Now()
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}

	var id int64
	err = r.pool.QueryRow(ctx, `
		insert into agent_queue (kind, payload, status, priority, max_attempts, run_after, dedupe_key)
		values ($1, $2::jsonb, $3, $4, $5, $6, nullif($7, ''))
		on conflict do nothing
		returning id
	`, kind, string(data), status, priority, maxAttempts, runAfter, dedupeKey).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("enqueue: %w", err)
	}
	return id, nil
}

func (r *Repository) ApproveTask(ctx context.Context, id int64) error {
	ct, err := r.pool.Exec(ctx, `
		update agent_queue
		set status = $2,
		    updated_at = now(),
		    run_after = case when run_after < now() then now() else run_after end,
		    last_error = null
		where id = $1 and status = $3
	`, id, StatusQueued, StatusWaitingApproval)
	if err != nil {
		return fmt.Errorf("approve task: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("approve task: task %d is not in %q", id, StatusWaitingApproval)
	}
	return nil
}

func (r *Repository) RejectTask(ctx context.Context, id int64, reason string) error {
	if reason == "" {
		reason = "rejected by reviewer"
	}
	ct, err := r.pool.Exec(ctx, `
		update agent_queue
		set status = $2,
		    updated_at = now(),
		    finished_at = now(),
		    last_error = $4
		where id = $1 and status = $3
	`, id, StatusRejected, StatusWaitingApproval, reason)
	if err != nil {
		return fmt.Errorf("reject task: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("reject task: task %d is not in %q", id, StatusWaitingApproval)
	}
	return nil
}

func (r *Repository) ClaimNext(ctx context.Context, worker string) (*Task, error) {
	if worker == "" {
		worker = "agent-worker"
	}
	row := r.pool.QueryRow(ctx, `
		with candidate as (
			select id
			from agent_queue
			where status = 'queued' and run_after <= now()
			order by priority desc, run_after asc, id asc
			for update skip locked
			limit 1
		)
		update agent_queue q
		set status = 'in_progress',
		    attempts = q.attempts + 1,
		    locked_by = $1,
		    locked_at = now(),
		    updated_at = now()
		from candidate
		where q.id = candidate.id
		returning q.id, q.kind, q.payload::text, q.status, q.priority, q.attempts, q.max_attempts,
		          coalesce(q.dedupe_key, ''), q.run_after, coalesce(q.locked_by, ''), q.created_at, q.updated_at
	`, worker)

	var t Task
	var payload string
	if err := row.Scan(&t.ID, &t.Kind, &payload, &t.Status, &t.Priority, &t.Attempts, &t.MaxAttempts, &t.DedupeKey, &t.RunAfter, &t.LockedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("claim next: %w", err)
	}
	t.Payload = []byte(payload)
	return &t, nil
}

func (r *Repository) MarkDone(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `
		update agent_queue
		set status = 'done',
		    finished_at = now(),
		    updated_at = now(),
		    last_error = null
		where id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("mark done: %w", err)
	}
	return nil
}

func (r *Repository) MarkFailed(ctx context.Context, id int64, errText string, retryAfter time.Duration) error {
	if errText == "" {
		errText = "unknown error"
	}

	status := StatusFailed
	var nextRun any = nil
	if retryAfter > 0 {
		status = StatusQueued
		nextRun = time.Now().Add(retryAfter)
	}

	_, err := r.pool.Exec(ctx, `
		update agent_queue
		set status = $2,
		    run_after = coalesce($3, run_after),
		    locked_by = null,
		    locked_at = null,
		    last_error = $4,
		    updated_at = now(),
		    finished_at = case when $2 = 'failed' then now() else null end
		where id = $1
	`, id, status, nextRun, errText)
	if err != nil {
		return fmt.Errorf("mark failed: %w", err)
	}
	return nil
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
