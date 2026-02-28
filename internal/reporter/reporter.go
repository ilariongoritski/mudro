package reporter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Summary struct {
	Text string
}

func BuildSummary(repoRoot string) (Summary, error) {
	mem, _ := readJSON(filepath.Join(repoRoot, ".codex", "memory.json"))
	runtime, _ := readJSON(filepath.Join(repoRoot, ".codex", "time_runtime.json"))
	top10 := readTop10(filepath.Join(repoRoot, ".codex", "top10.md"), 3)
	todo := readTodo(filepath.Join(repoRoot, ".codex", "todo.md"), 3)
	last := lastRuns(repoRoot, 3)

	var b strings.Builder
	b.WriteString("Mudro reporter\n")
	b.WriteString(time.Now().Format("2006-01-02 15:04:05") + "\n")

	if daySec, totalSec, runs := parseMemory(mem); runs > 0 {
		b.WriteString(fmt.Sprintf("Время работы: сегодня %s, всего %s, прогонов %d\n", fmtDuration(daySec), fmtDuration(totalSec), runs))
	}
	if resp, totalMS := parseRuntime(runtime); resp > 0 {
		b.WriteString(fmt.Sprintf("Генерация ответов: %s, ответов %d\n", fmtDuration(totalMS/1000), resp))
	}
	if len(top10) > 0 {
		b.WriteString("Top-3 изменения:\n")
		for _, t := range top10 {
			b.WriteString("- " + t + "\n")
		}
	}
	if len(todo) > 0 {
		b.WriteString("Фокус TODO:\n")
		for _, t := range todo {
			b.WriteString("- " + t + "\n")
		}
	}
	if len(last) > 0 {
		b.WriteString("Последние прогоны:\n")
		for _, r := range last {
			b.WriteString("- " + r + "\n")
		}
	}
	if q, e, err := loadAgentMetrics(); err == nil {
		b.WriteString("Агент (24ч):\n")
		b.WriteString(fmt.Sprintf("- created/done/failed/retry: %d/%d/%d/%d\n", e.Created, e.Done, e.Failed, e.Retry))
		b.WriteString(fmt.Sprintf("- approved/rejected: %d/%d\n", e.Approved, e.Rejected))
		b.WriteString(fmt.Sprintf("- queue queued/waiting/in_progress: %d/%d/%d\n", q.Queued, q.WaitingApproval, q.InProgress))
		if len(e.TopKinds) > 0 {
			b.WriteString("- частые task kinds:\n")
			for _, k := range e.TopKinds {
				b.WriteString(fmt.Sprintf("  - %s: %d\n", k.Kind, k.Count))
			}
		}
	}
	return Summary{Text: strings.TrimSpace(b.String())}, nil
}

type agentQueueStatus struct {
	Queued          int64
	WaitingApproval int64
	InProgress      int64
}

type topKind struct {
	Kind  string
	Count int64
}

type agentEventStats struct {
	Created  int64
	Done     int64
	Failed   int64
	Retry    int64
	Approved int64
	Rejected int64
	TopKinds []topKind
}

func loadAgentMetrics() (agentQueueStatus, agentEventStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, config.DSN())
	if err != nil {
		return agentQueueStatus{}, agentEventStats{}, err
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return agentQueueStatus{}, agentEventStats{}, err
	}

	var q agentQueueStatus
	rows, err := pool.Query(ctx, `select status, count(*) from agent_queue group by status`)
	if err != nil {
		return agentQueueStatus{}, agentEventStats{}, err
	}
	for rows.Next() {
		var st string
		var cnt int64
		if err := rows.Scan(&st, &cnt); err != nil {
			rows.Close()
			return agentQueueStatus{}, agentEventStats{}, err
		}
		switch st {
		case "queued":
			q.Queued = cnt
		case "waiting_approval":
			q.WaitingApproval = cnt
		case "in_progress":
			q.InProgress = cnt
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return agentQueueStatus{}, agentEventStats{}, err
	}
	rows.Close()

	var e agentEventStats
	err = pool.QueryRow(ctx, `
		select
			count(*) filter (where event_type='task.created'),
			count(*) filter (where event_type='task.done'),
			count(*) filter (where event_type='task.failed'),
			count(*) filter (where event_type='task.retry_scheduled'),
			count(*) filter (where event_type='task.approved'),
			count(*) filter (where event_type='task.rejected')
		from agent_task_events
		where occurred_at >= now() - interval '24 hours'
	`).Scan(&e.Created, &e.Done, &e.Failed, &e.Retry, &e.Approved, &e.Rejected)
	if err != nil {
		return agentQueueStatus{}, agentEventStats{}, err
	}

	krows, err := pool.Query(ctx, `
		select coalesce(kind,'unknown') as kind, count(*) as cnt
		from agent_task_events
		where occurred_at >= now() - interval '24 hours'
		  and event_type in ('task.done','task.failed','task.retry_scheduled')
		group by 1
		order by cnt desc, kind asc
		limit 3
	`)
	if err == nil {
		defer krows.Close()
		for krows.Next() {
			var k topKind
			if err := krows.Scan(&k.Kind, &k.Count); err != nil {
				break
			}
			e.TopKinds = append(e.TopKinds, k)
		}
	}

	return q, e, nil
}

func ResolveChatID(repoRoot string, envChatID int64) int64 {
	if envChatID > 0 {
		return envChatID
	}
	path := filepath.Join(repoRoot, ".codex", "tg_control.jsonl")
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	type event struct {
		ChatID int64 `json:"chat_id"`
	}
	var last int64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e event
		if json.Unmarshal(sc.Bytes(), &e) == nil && e.ChatID > 0 {
			last = e.ChatID
		}
	}
	return last
}

func lastRuns(repoRoot string, n int) []string {
	logDir := filepath.Join(repoRoot, ".codex", "logs")
	ents, err := os.ReadDir(logDir)
	if err != nil {
		return nil
	}
	ids := make([]string, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	sort.Strings(ids)
	if len(ids) > n {
		ids = ids[len(ids)-n:]
	}
	return ids
}

func readTop10(path string, n int) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(b), "\n")
	out := make([]string, 0, n)
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "1.") || strings.HasPrefix(l, "2.") || strings.HasPrefix(l, "3.") || strings.HasPrefix(l, "4.") || strings.HasPrefix(l, "5.") {
			out = append(out, trimAfter(l, 120))
			if len(out) >= n {
				break
			}
		}
	}
	return out
}

func readTodo(path string, n int) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(b), "\n")
	out := make([]string, 0, n)
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "- [ ] ") {
			out = append(out, trimAfter(strings.TrimPrefix(l, "- [ ] "), 120))
			if len(out) >= n {
				break
			}
		}
	}
	return out
}

func readJSON(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func parseMemory(mem map[string]any) (todaySec, totalSec int64, runs int64) {
	if mem == nil {
		return
	}
	if totals, ok := mem["totals"].(map[string]any); ok {
		totalSec = toInt64(totals["total_seconds"])
		runs = toInt64(totals["runs"])
	}
	today := time.Now().Format("2006-01-02")
	if days, ok := mem["days"].(map[string]any); ok {
		if d, ok := days[today].(map[string]any); ok {
			todaySec = toInt64(d["total_seconds"])
		}
	}
	return
}

func parseRuntime(mem map[string]any) (responses int64, totalMS int64) {
	if mem == nil {
		return
	}
	if totals, ok := mem["totals"].(map[string]any); ok {
		responses = toInt64(totals["responses"])
		totalMS = toInt64(totals["total_ms"])
	}
	return
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	default:
		return 0
	}
}

func fmtDuration(sec int64) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func trimAfter(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n]) + "..."
}
