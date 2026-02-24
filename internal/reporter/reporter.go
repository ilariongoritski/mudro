package reporter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
	return Summary{Text: strings.TrimSpace(b.String())}, nil
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
