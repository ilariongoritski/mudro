package repository

import "github.com/goritskimihail/mudro/internal/agent/domain"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/goritskimihail/mudro/internal/events"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRepository struct {
	pool      *pgxpool.Pool
	publisher events.Publisher
}

func NewPgRepository(pool *pgxpool.Pool) TaskRepository {
	return &pgRepository{pool: pool, publisher: events.NoopPublisher{}}
}

func NewPgRepositoryWithPublisher(pool *pgxpool.Pool, p events.Publisher) TaskRepository {
	if p == nil {
		p = events.NoopPublisher{}
	}
	return &pgRepository{pool: pool, publisher: p}
}

func (r *pgRepository) Enqueue(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
	return r.enqueueWithStatus(ctx, kind, payload, priority, runAfter, maxAttempts, dedupeKey, domain.StatusQueued)
}

func (r *pgRepository) EnqueueWaitingApproval(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
	return r.enqueueWithStatus(ctx, kind, payload, priority, runAfter, maxAttempts, dedupeKey, domain.StatusWaitingApproval)
}

func (r *pgRepository) enqueueWithStatus(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey, status string) (int64, error) {
	if kind == "" {
		return 0, errors.New("kind is required")
	}
	if status == "" {
		status = domain.StatusQueued
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
	r.publishTaskEvent(ctx, events.TaskEvent{
		EventType: "task.created",
		TaskID:    id,
		Kind:      kind,
		Status:    status,
		DedupeKey: dedupeKey,
		Occurred:  time.Now().UTC(),
	})
	return id, nil
}

func (r *pgRepository) ApproveTask(ctx context.Context, id int64) error {
	var kind, dedupeKey string
	err := r.pool.QueryRow(ctx, `
		update agent_queue
		set status = $2,
		    updated_at = now(),
		    run_after = case when run_after < now() then now() else run_after end,
		    last_error = null
		where id = $1 and status = $3
		returning kind, coalesce(dedupe_key, '')
	`, id, domain.StatusQueued, domain.StatusWaitingApproval).Scan(&kind, &dedupeKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("approve task: task %d is not in %q", id, domain.StatusWaitingApproval)
		}
		return fmt.Errorf("approve task: %w", err)
	}
	r.publishTaskEvent(ctx, events.TaskEvent{
		EventType: "task.approved",
		TaskID:    id,
		Kind:      kind,
		Status:    domain.StatusQueued,
		DedupeKey: dedupeKey,
		Occurred:  time.Now().UTC(),
	})
	return nil
}

func (r *pgRepository) RejectTask(ctx context.Context, id int64, reason string) error {
	if reason == "" {
		reason = "rejected by reviewer"
	}
	var kind, dedupeKey string
	err := r.pool.QueryRow(ctx, `
		update agent_queue
		set status = $2,
		    updated_at = now(),
		    finished_at = now(),
		    last_error = $4
		where id = $1 and status = $3
		returning kind, coalesce(dedupe_key, '')
	`, id, domain.StatusRejected, domain.StatusWaitingApproval, reason).Scan(&kind, &dedupeKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("reject task: task %d is not in %q", id, domain.StatusWaitingApproval)
		}
		return fmt.Errorf("reject task: %w", err)
	}
	r.publishTaskEvent(ctx, events.TaskEvent{
		EventType: "task.rejected",
		TaskID:    id,
		Kind:      kind,
		Status:    domain.StatusRejected,
		DedupeKey: dedupeKey,
		Error:     reason,
		Occurred:  time.Now().UTC(),
	})
	return nil
}

func (r *pgRepository) ClaimNext(ctx context.Context, worker string) (*domain.Task, error) {
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

	var t domain.Task
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

func (r *pgRepository) MarkDone(ctx context.Context, id int64) error {
	var kind, dedupeKey string
	err := r.pool.QueryRow(ctx, `
		update agent_queue
		set status = 'done',
		    finished_at = now(),
		    updated_at = now(),
		    last_error = null
		where id = $1
		returning kind, coalesce(dedupe_key, '')
	`, id).Scan(&kind, &dedupeKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("mark done: task %d not found", id)
		}
		return fmt.Errorf("mark done: %w", err)
	}
	r.publishTaskEvent(ctx, events.TaskEvent{
		EventType: "task.done",
		TaskID:    id,
		Kind:      kind,
		Status:    domain.StatusDone,
		DedupeKey: dedupeKey,
		Occurred:  time.Now().UTC(),
	})
	return nil
}

func (r *pgRepository) MarkFailed(ctx context.Context, id int64, errText string, retryAfter time.Duration) error {
	if errText == "" {
		errText = "unknown error"
	}

	status := domain.StatusFailed
	var nextRun any = nil
	if retryAfter > 0 {
		status = domain.StatusQueued
		nextRun = time.Now().Add(retryAfter)
	}

	var kind, dedupeKey string
	err := r.pool.QueryRow(ctx, `
		update agent_queue
		set status = $2,
		    run_after = coalesce($3, run_after),
		    locked_by = null,
		    locked_at = null,
		    last_error = $4,
		    updated_at = now(),
		    finished_at = case when $2 = 'failed' then now() else null end
		where id = $1
		returning kind, coalesce(dedupe_key, '')
	`, id, status, nextRun, errText).Scan(&kind, &dedupeKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("mark failed: task %d not found", id)
		}
		return fmt.Errorf("mark failed: %w", err)
	}
	eventType := "task.failed"
	if status == domain.StatusQueued {
		eventType = "task.retry_scheduled"
	}
	r.publishTaskEvent(ctx, events.TaskEvent{
		EventType: eventType,
		TaskID:    id,
		Kind:      kind,
		Status:    status,
		DedupeKey: dedupeKey,
		Error:     errText,
		Occurred:  time.Now().UTC(),
	})
	return nil
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func (r *pgRepository) publishTaskEvent(ctx context.Context, ev events.TaskEvent) {
	if ev.Occurred.IsZero() {
		ev.Occurred = time.Now().UTC()
	}
	if ev.EventID == "" {
		ev.EventID = fmt.Sprintf("%d-%s-%d", ev.TaskID, ev.EventType, ev.Occurred.UnixNano())
	}

	if err := r.persistTaskEvent(ctx, ev); err != nil {
		log.Printf("persist task event %s task_id=%d: %v", ev.EventType, ev.TaskID, err)
	}

	if r.publisher == nil {
		return
	}
	if err := r.publisher.PublishTaskEvent(ctx, ev); err != nil {
		log.Printf("publish task event %s task_id=%d: %v", ev.EventType, ev.TaskID, err)
	}
}

func (r *pgRepository) persistTaskEvent(ctx context.Context, ev events.TaskEvent) error {
	if r.pool == nil {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
		insert into agent_task_events (
			event_id, task_id, event_type, status, kind, dedupe_key, error, occurred_at
		) values ($1,$2,$3,$4,nullif($5,''),nullif($6,''),nullif($7,''),$8)
		on conflict (event_id) do nothing
	`, ev.EventID, ev.TaskID, ev.EventType, ev.Status, ev.Kind, ev.DedupeKey, ev.Error, ev.Occurred)
	return err
}

