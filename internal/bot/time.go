package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
)

type memoryJSON struct {
	Version   int                    `json:"version"`
	UpdatedAt string                 `json:"updated_at"`
	Days      map[string]dayWorkStat `json:"days"`
	Totals    totalsStat             `json:"totals"`
}

type dayWorkStat struct {
	Runs         int               `json:"runs"`
	TotalSeconds int64             `json:"total_seconds"`
	Entries      []dayRunEntryStat `json:"entries"`
}

type dayRunEntryStat struct {
	RunID            string `json:"run_id"`
	StartedAt        string `json:"started_at"`
	EstimatedSeconds int64  `json:"estimated_seconds"`
}

type totalsStat struct {
	Runs         int   `json:"runs"`
	TotalSeconds int64 `json:"total_seconds"`
}

type runRef struct {
	id string
	ts time.Time
}

func (r *Runner) TimeSummary() ([]byte, error) {
	mem, err := r.rebuildMemoryJSON()
	if err != nil {
		return nil, err
	}
	today := time.Now().Format("2006-01-02")
	todayStat := mem.Days[today]

	var out strings.Builder
	out.WriteString("Время работы:\n")
	out.WriteString(fmt.Sprintf("- Сегодня: %s (%d мин, %d сек), прогонов: %d\n",
		fmtDuration(todayStat.TotalSeconds), todayStat.TotalSeconds/60, todayStat.TotalSeconds, todayStat.Runs))
	out.WriteString(fmt.Sprintf("- Всего: %s (%d мин, %d сек), прогонов: %d\n",
		fmtDuration(mem.Totals.TotalSeconds), mem.Totals.TotalSeconds/60, mem.Totals.TotalSeconds, mem.Totals.Runs))

	if rt, err := readRuntimeTime(filepath.Join(r.RepoRoot, ".codex", "time_runtime.json")); err == nil && rt != nil && rt.Totals.Responses > 0 {
		avgMS := rt.Totals.TotalMS / int64(rt.Totals.Responses)
		out.WriteString(fmt.Sprintf("- Генерация ответов: %s (%d мин, %d сек), ответов: %d, среднее: %d мс\n",
			fmtDuration(rt.Totals.TotalMS/1000), (rt.Totals.TotalMS/1000)/60, rt.Totals.TotalMS/1000,
			rt.Totals.Responses, avgMS))
	}
	out.WriteString("- Память JSON: .codex/memory.json")
	return []byte(out.String()), nil
}

func (r *Runner) rebuildMemoryJSON() (*memoryJSON, error) {
	logDir := filepath.Join(r.RepoRoot, config.CodexLogsDir())
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("read logs dir: %w", err)
	}

	runs := make([]runRef, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ts, err := time.Parse("20060102-1504", e.Name())
		if err != nil {
			continue
		}
		runs = append(runs, runRef{id: e.Name(), ts: ts})
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].ts.Before(runs[j].ts) })

	mem := &memoryJSON{
		Version:   1,
		UpdatedAt: time.Now().Format(time.RFC3339),
		Days:      make(map[string]dayWorkStat),
	}

	for i := range runs {
		cur := runs[i]
		est := estimateRunSeconds(cur, runs, i)
		day := cur.ts.Format("2006-01-02")

		stat := mem.Days[day]
		stat.Runs++
		stat.TotalSeconds += est
		stat.Entries = append(stat.Entries, dayRunEntryStat{
			RunID:            cur.id,
			StartedAt:        cur.ts.Format(time.RFC3339),
			EstimatedSeconds: est,
		})
		mem.Days[day] = stat

		mem.Totals.Runs++
		mem.Totals.TotalSeconds += est
	}

	path := filepath.Join(r.RepoRoot, ".codex", "memory.json")
	b, err := json.MarshalIndent(mem, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal memory json: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return nil, fmt.Errorf("write memory json: %w", err)
	}
	return mem, nil
}

func estimateRunSeconds(cur runRef, runs []runRef, idx int) int64 {
	const (
		minSec     = int64(120)  // 2 min
		maxSec     = int64(3600) // 60 min
		defaultSec = int64(600)  // 10 min
	)
	if idx+1 >= len(runs) {
		return defaultSec
	}
	next := runs[idx+1]
	if next.ts.Format("2006-01-02") != cur.ts.Format("2006-01-02") {
		return defaultSec
	}
	diff := int64(next.ts.Sub(cur.ts).Seconds())
	if diff < minSec {
		return minSec
	}
	if diff > maxSec {
		return defaultSec
	}
	return diff
}

func fmtDuration(sec int64) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
