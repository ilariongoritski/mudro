package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type agentQueueSnapshot struct {
	Queued          int64
	WaitingApproval int64
	InProgress      int64
}

type kindCount struct {
	Kind  string
	Count int64
}

type agentEventSnapshot struct {
	Created  int64
	Done     int64
	Failed   int64
	Retry    int64
	Approved int64
	Rejected int64
	TopKinds []kindCount
}

func (r *Runner) Agent24Summary() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, r.DSN)
	if err != nil {
		return nil, fmt.Errorf("подключение к БД: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping БД: %w", err)
	}

	now := time.Now()
	last24, err := readAgentEventWindow(ctx, pool, now.Add(-24*time.Hour), now)
	if err != nil {
		return nil, fmt.Errorf("метрики 24ч: %w", err)
	}
	prev24, err := readAgentEventWindow(ctx, pool, now.Add(-48*time.Hour), now.Add(-24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("метрики предыдущих 24ч: %w", err)
	}
	queue, err := readAgentQueueSnapshot(ctx, pool)
	if err != nil {
		return nil, fmt.Errorf("снимок очереди: %w", err)
	}

	var b strings.Builder
	b.WriteString("Агент за последние 24 часа:\n")
	b.WriteString(fmt.Sprintf("- Создано задач: %d\n", last24.Created))
	b.WriteString(fmt.Sprintf("- Завершено успешно: %d\n", last24.Done))
	b.WriteString(fmt.Sprintf("- Ошибок: %d\n", last24.Failed))
	b.WriteString(fmt.Sprintf("- Ретраев запланировано: %d\n", last24.Retry))
	b.WriteString(fmt.Sprintf("- Аппрув/реджект: %d/%d\n", last24.Approved, last24.Rejected))

	successDenom := last24.Done + last24.Failed
	if successDenom > 0 {
		successRate := float64(last24.Done) * 100.0 / float64(successDenom)
		b.WriteString(fmt.Sprintf("- Успешность выполнения: %.1f%%\n", successRate))
	}

	b.WriteString("Что изменилось относительно предыдущих 24 часов:\n")
	b.WriteString("- Завершено: " + signedDelta(last24.Done-prev24.Done) + "\n")
	b.WriteString("- Ошибок: " + signedDelta(last24.Failed-prev24.Failed) + "\n")
	b.WriteString("- Ретраев: " + signedDelta(last24.Retry-prev24.Retry) + "\n")

	b.WriteString("Очередь сейчас:\n")
	b.WriteString(fmt.Sprintf("- queued: %d\n", queue.Queued))
	b.WriteString(fmt.Sprintf("- waiting_approval: %d\n", queue.WaitingApproval))
	b.WriteString(fmt.Sprintf("- in_progress: %d\n", queue.InProgress))

	if len(last24.TopKinds) > 0 {
		b.WriteString("Чаще всего агент делал:\n")
		for _, k := range last24.TopKinds {
			b.WriteString(fmt.Sprintf("- %s: %d\n", k.Kind, k.Count))
		}
	}

	b.WriteString("Простой вывод:\n")
	switch {
	case last24.Failed > 0 && last24.Failed >= last24.Done:
		b.WriteString("- Нужна стабилизация: ошибок не меньше, чем успешных выполнений.\n")
	case queue.WaitingApproval > 0:
		b.WriteString("- Есть задачи, которые ждут ручного решения (approve/reject).\n")
	default:
		b.WriteString("- Контур выглядит стабильным, критичных сигналов нет.\n")
	}

	return []byte(strings.TrimSpace(b.String())), nil
}

func readAgentQueueSnapshot(ctx context.Context, pool *pgxpool.Pool) (agentQueueSnapshot, error) {
	var q agentQueueSnapshot
	rows, err := pool.Query(ctx, `select status, count(*) from agent_queue group by status`)
	if err != nil {
		return q, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var cnt int64
		if err := rows.Scan(&status, &cnt); err != nil {
			return q, err
		}
		switch status {
		case "queued":
			q.Queued = cnt
		case "waiting_approval":
			q.WaitingApproval = cnt
		case "in_progress":
			q.InProgress = cnt
		}
	}
	return q, rows.Err()
}

func readAgentEventWindow(ctx context.Context, pool *pgxpool.Pool, from, to time.Time) (agentEventSnapshot, error) {
	var e agentEventSnapshot
	err := pool.QueryRow(ctx, `
		select
			count(*) filter (where event_type='task.created'),
			count(*) filter (where event_type='task.done'),
			count(*) filter (where event_type='task.failed'),
			count(*) filter (where event_type='task.retry_scheduled'),
			count(*) filter (where event_type='task.approved'),
			count(*) filter (where event_type='task.rejected')
		from agent_task_events
		where occurred_at >= $1 and occurred_at < $2
	`, from, to).Scan(&e.Created, &e.Done, &e.Failed, &e.Retry, &e.Approved, &e.Rejected)
	if err != nil {
		return e, err
	}

	rows, err := pool.Query(ctx, `
		select coalesce(kind, 'unknown') as kind, count(*) as cnt
		from agent_task_events
		where occurred_at >= $1 and occurred_at < $2
		  and event_type in ('task.done','task.failed','task.retry_scheduled')
		group by 1
		order by cnt desc, kind asc
		limit 5
	`, from, to)
	if err != nil {
		return e, nil
	}
	defer rows.Close()
	for rows.Next() {
		var k kindCount
		if err := rows.Scan(&k.Kind, &k.Count); err != nil {
			return e, nil
		}
		e.TopKinds = append(e.TopKinds, k)
	}
	return e, nil
}

func signedDelta(v int64) string {
	if v > 0 {
		return fmt.Sprintf("+%d", v)
	}
	return fmt.Sprintf("%d", v)
}
