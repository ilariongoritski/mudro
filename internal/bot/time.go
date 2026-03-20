package bot

import (
	"encoding/json"
	"fmt"
	"math"
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

	runtimePath := filepath.Join(r.RepoRoot, ".codex", "time_runtime.json")
	if rt, err := readRuntimeTime(runtimePath); err == nil && rt != nil && rt.Totals.Responses > 0 {
		avgMS := rt.Totals.TotalMS / int64(rt.Totals.Responses)
		thinkSec := rt.Totals.ProcessMS / 1000
		evapProcessMl := estimateEvaporationML(thinkSec)
		evapTotalMl := estimateEvaporationML(rt.Totals.TotalMS / 1000)
		out.WriteString("- Учёт runtime ответов бота (факт):\n")
		out.WriteString(fmt.Sprintf("  - total: %s (%d мс), ответов: %d, среднее: %d мс, max: %d мс\n",
			fmtDuration(rt.Totals.TotalMS/1000), rt.Totals.TotalMS, rt.Totals.Responses, avgMS, rt.Totals.MaxTotalMS))
		out.WriteString(fmt.Sprintf("  - process: %s (%d мс)\n", fmtDuration(rt.Totals.ProcessMS/1000), rt.Totals.ProcessMS))
		out.WriteString(fmt.Sprintf("  - send: %s (%d мс)\n", fmtDuration(rt.Totals.SendMS/1000), rt.Totals.SendMS))
		out.WriteString(fmt.Sprintf("  - unknown: %s (%d мс)\n", fmtDuration(rt.Totals.UnknownMS/1000), rt.Totals.UnknownMS))
		out.WriteString(fmt.Sprintf("- Потрачено на обдумывание/обработку (process): %s.\n", formatHoursMinutes(thinkSec)))
		out.WriteString(fmt.Sprintf("- Вода (оценка): process=%d мл, total=%d мл.\n", evapProcessMl, evapTotalMl))
		out.WriteString("- Модель воды: 50 мл/час, расчет оценочный.\n")
		out.WriteString("- Детализация по командам (top 5 по total):\n")
		for _, line := range topRuntimeCommands(rt, 5) {
			out.WriteString("  - " + line + "\n")
		}
		llmTotalMS, llmProcessMS, llmResponses := llmRuntimeTotals(rt)
		if llmResponses > 0 {
			llmAvgMS := llmTotalMS / int64(llmResponses)
			out.WriteString("- LLM runtime (только ответы модели /mudro + chat-mode):\n")
			out.WriteString(fmt.Sprintf("  - total: %s (%d мс), ответов: %d, среднее: %d мс\n",
				fmtDuration(llmTotalMS/1000), llmTotalMS, llmResponses, llmAvgMS))
			out.WriteString(fmt.Sprintf("  - process: %s (%d мс)\n", fmtDuration(llmProcessMS/1000), llmProcessMS))
		}
	}
	if agg, err := buildRuntimeAggregate(runtimePath); err == nil && agg.HasExternal {
		out.WriteString("- Общий runtime (бот + backfill из чата/кловбота):\n")
		out.WriteString(fmt.Sprintf("  - total: %s (%d мс), ответов/итераций: %d\n",
			fmtDuration(agg.TotalMS/1000), agg.TotalMS, agg.Responses))
		for _, src := range agg.External {
			out.WriteString(fmt.Sprintf("  - source %s: %s (%d мс), count=%d\n",
				src.Name, fmtDuration(src.TotalMS/1000), src.TotalMS, src.Responses))
		}
	}
	out.WriteString("- Память JSON: .codex/memory.json, .codex/time_runtime.json")
	return []byte(out.String()), nil
}

func topRuntimeCommands(rt *runtimeTimeMemory, n int) []string {
	type item struct {
		cmd  string
		stat runtimeCommand
	}
	items := make([]item, 0, len(rt.ByCommand))
	for cmd, st := range rt.ByCommand {
		items = append(items, item{cmd: cmd, stat: st})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].stat.TotalMS == items[j].stat.TotalMS {
			return items[i].cmd < items[j].cmd
		}
		return items[i].stat.TotalMS > items[j].stat.TotalMS
	})
	if n <= 0 || n > len(items) {
		n = len(items)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		st := items[i].stat
		out = append(out, fmt.Sprintf("/%s: total=%dms, process=%dms, send=%dms, responses=%d",
			items[i].cmd, st.TotalMS, st.ProcessMS, st.SendMS, st.Responses))
	}
	return out
}

type runtimeExternalSource struct {
	Name      string
	TotalMS   int64
	Responses int
}

type runtimeAggregate struct {
	TotalMS     int64
	Responses   int
	External    []runtimeExternalSource
	HasExternal bool
}

func buildRuntimeAggregate(path string) (*runtimeAggregate, error) {
	rt, extras, err := readRuntimeTimeWithExtras(path)
	if err != nil || rt == nil {
		return nil, err
	}
	agg := &runtimeAggregate{
		TotalMS:   rt.Totals.TotalMS,
		Responses: rt.Totals.Responses,
	}

	for name, raw := range extras {
		meta := struct {
			TotalMS   *int64 `json:"total_ms"`
			Responses *int   `json:"responses"`
			Turns     *int   `json:"turns"`
		}{}
		if err := json.Unmarshal(raw, &meta); err != nil || meta.TotalMS == nil || *meta.TotalMS <= 0 {
			continue
		}

		count := 0
		switch {
		case meta.Responses != nil:
			count = *meta.Responses
		case meta.Turns != nil:
			count = *meta.Turns
		}

		agg.External = append(agg.External, runtimeExternalSource{
			Name:      name,
			TotalMS:   *meta.TotalMS,
			Responses: count,
		})
		agg.TotalMS += *meta.TotalMS
		agg.Responses += count
	}

	sort.Slice(agg.External, func(i, j int) bool { return agg.External[i].TotalMS > agg.External[j].TotalMS })
	agg.HasExternal = len(agg.External) > 0
	return agg, nil
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

func formatHoursMinutes(sec int64) string {
	if sec < 0 {
		sec = 0
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	return fmt.Sprintf("%d часов %d минут", h, m)
}

func estimateEvaporationML(sec int64) int64 {
	// Approximation: 50 ml/hour "evaporation" during thinking time.
	const mlPerHour = 50.0
	ml := (float64(sec) / 3600.0) * mlPerHour
	if ml < 0 {
		return 0
	}
	return int64(math.Round(ml))
}

func llmRuntimeTotals(rt *runtimeTimeMemory) (totalMS int64, processMS int64, responses int) {
	for cmd, st := range rt.ByCommand {
		if cmd == "mudro" || cmd == "mudro_chat" {
			totalMS += st.TotalMS
			processMS += st.ProcessMS
			responses += st.Responses
		}
	}
	return totalMS, processMS, responses
}
