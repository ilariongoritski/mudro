package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type runtimeTimeMemory struct {
	Version   int                       `json:"version"`
	UpdatedAt string                    `json:"updated_at"`
	Totals    runtimeTotals             `json:"totals"`
	ByCommand map[string]runtimeCommand `json:"by_command"`
}

type runtimeTotals struct {
	Responses  int   `json:"responses"`
	TotalMS    int64 `json:"total_ms"`
	ProcessMS  int64 `json:"process_ms"`
	SendMS     int64 `json:"send_ms"`
	UnknownMS  int64 `json:"unknown_ms"`
	MaxTotalMS int64 `json:"max_total_ms"`
	LastTotal  int64 `json:"last_total_ms"`
}

type runtimeCommand struct {
	Responses  int   `json:"responses"`
	TotalMS    int64 `json:"total_ms"`
	ProcessMS  int64 `json:"process_ms"`
	SendMS     int64 `json:"send_ms"`
	UnknownMS  int64 `json:"unknown_ms"`
	MaxTotalMS int64 `json:"max_total_ms"`
	LastTotal  int64 `json:"last_total_ms"`
}

var commandStartStore = struct {
	mu sync.Mutex
	m  map[string]time.Time
}{
	m: make(map[string]time.Time),
}

func markCommandStart(updateKey string) {
	commandStartStore.mu.Lock()
	defer commandStartStore.mu.Unlock()
	commandStartStore.m[updateKey] = time.Now()
}

func popCommandStart(updateKey string) (time.Time, bool) {
	commandStartStore.mu.Lock()
	defer commandStartStore.mu.Unlock()
	t, ok := commandStartStore.m[updateKey]
	if ok {
		delete(commandStartStore.m, updateKey)
	}
	return t, ok
}

func updateKey(chatID int64, msgID int) string {
	return fmt.Sprintf("%d:%d", chatID, msgID)
}

func recordRuntimeTime(repoRoot string, command string, process, send, total time.Duration) error {
	path := filepath.Join(repoRoot, ".codex", "time_runtime.json")
	mem, extras, _ := readRuntimeTimeWithExtras(path)
	if mem == nil {
		mem = &runtimeTimeMemory{
			Version:   1,
			ByCommand: make(map[string]runtimeCommand),
		}
	}

	totalMS := nonNegativeMS(total)
	processMS := nonNegativeMS(process)
	sendMS := nonNegativeMS(send)
	unknownMS := totalMS - processMS - sendMS
	if unknownMS < 0 {
		unknownMS = 0
	}

	mem.UpdatedAt = time.Now().Format(time.RFC3339)
	mem.Totals.Responses++
	mem.Totals.TotalMS += totalMS
	mem.Totals.ProcessMS += processMS
	mem.Totals.SendMS += sendMS
	mem.Totals.UnknownMS += unknownMS
	mem.Totals.LastTotal = totalMS
	if totalMS > mem.Totals.MaxTotalMS {
		mem.Totals.MaxTotalMS = totalMS
	}

	command = strings.TrimSpace(strings.TrimPrefix(command, "/"))
	if command == "" {
		command = "unknown"
	}
	cs := mem.ByCommand[command]
	cs.Responses++
	cs.TotalMS += totalMS
	cs.ProcessMS += processMS
	cs.SendMS += sendMS
	cs.UnknownMS += unknownMS
	cs.LastTotal = totalMS
	if totalMS > cs.MaxTotalMS {
		cs.MaxTotalMS = totalMS
	}
	mem.ByCommand[command] = cs

	return writeRuntimeTime(path, mem, extras)
}

func nonNegativeMS(d time.Duration) int64 {
	ms := d.Milliseconds()
	if ms < 0 {
		return 0
	}
	return ms
}

func readRuntimeTime(path string) (*runtimeTimeMemory, error) {
	mem, _, err := readRuntimeTimeWithExtras(path)
	return mem, err
}

func readRuntimeTimeWithExtras(path string) (*runtimeTimeMemory, map[string]json.RawMessage, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, err
	}

	var mem runtimeTimeMemory
	if err := json.Unmarshal(b, &mem); err != nil {
		return nil, nil, err
	}
	if mem.ByCommand == nil {
		mem.ByCommand = make(map[string]runtimeCommand)
	}

	// Preserve any additional top-level sections (for example, desktop/chat backfills).
	delete(raw, "version")
	delete(raw, "updated_at")
	delete(raw, "totals")
	delete(raw, "by_command")
	return &mem, raw, nil
}

func writeRuntimeTime(path string, mem *runtimeTimeMemory, extras map[string]json.RawMessage) error {
	base, err := json.Marshal(mem)
	if err != nil {
		return err
	}
	out := make(map[string]any)
	if err := json.Unmarshal(base, &out); err != nil {
		return err
	}

	for k, v := range extras {
		if _, exists := out[k]; exists {
			continue
		}
		var decoded any
		if err := json.Unmarshal(v, &decoded); err != nil {
			continue
		}
		out[k] = decoded
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
