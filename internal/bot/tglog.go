package bot

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type tgControlEvent struct {
	TS        string `json:"ts"`
	Username  string `json:"username,omitempty"`
	ChatID    int64  `json:"chat_id"`
	Command   string `json:"command"`
	Args      string `json:"args,omitempty"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	ReplyHint string `json:"reply_hint,omitempty"`
}

func (r *Runner) TelegramControlLog() ([]byte, error) {
	path := filepath.Join(r.RepoRoot, ".codex", "tg_control.jsonl")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte("Лог управления из Telegram пока пуст"), nil
		}
		return nil, fmt.Errorf("open tg control log: %w", err)
	}
	defer f.Close()

	events := make([]tgControlEvent, 0, 64)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e tgControlEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		events = append(events, e)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan tg control log: %w", err)
	}
	if len(events) == 0 {
		return []byte("Лог управления из Telegram пока пуст"), nil
	}

	start := 0
	if len(events) > 20 {
		start = len(events) - 20
	}
	var out strings.Builder
	out.WriteString("Telegram управление (последние записи):\n")
	for _, e := range events[start:] {
		out.WriteString(fmt.Sprintf("- %s | @%s | %s %s | %s\n",
			shortTS(e.TS), nonEmpty(e.Username, "unknown"), e.Command, trimForLog(e.Args, 48), e.Status))
		if e.Error != "" {
			out.WriteString("  err: " + trimForLog(e.Error, 100) + "\n")
		}
	}
	return []byte(strings.TrimSpace(out.String())), nil
}

func appendTGControlEvent(repoRoot string, e tgControlEvent) error {
	path := filepath.Join(repoRoot, ".codex", "tg_control.jsonl")
	if e.TS == "" {
		e.TS = time.Now().Format(time.RFC3339)
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(string(b) + "\n")
	return err
}

func nonEmpty(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func trimForLog(s string, max int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}

func shortTS(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("2006-01-02 15:04:05")
}
