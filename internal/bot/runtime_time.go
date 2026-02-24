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
	Responses int   `json:"responses"`
	TotalMS   int64 `json:"total_ms"`
}

type runtimeCommand struct {
	Responses int   `json:"responses"`
	TotalMS   int64 `json:"total_ms"`
	MaxMS     int64 `json:"max_ms"`
	LastMS    int64 `json:"last_ms"`
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

func recordRuntimeTime(repoRoot string, command string, d time.Duration) error {
	path := filepath.Join(repoRoot, ".codex", "time_runtime.json")
	mem, _ := readRuntimeTime(path)
	if mem == nil {
		mem = &runtimeTimeMemory{
			Version:   1,
			ByCommand: make(map[string]runtimeCommand),
		}
	}

	ms := d.Milliseconds()
	if ms < 0 {
		ms = 0
	}
	mem.UpdatedAt = time.Now().Format(time.RFC3339)
	mem.Totals.Responses++
	mem.Totals.TotalMS += ms

	command = strings.TrimSpace(strings.TrimPrefix(command, "/"))
	if command == "" {
		command = "unknown"
	}
	cs := mem.ByCommand[command]
	cs.Responses++
	cs.TotalMS += ms
	cs.LastMS = ms
	if ms > cs.MaxMS {
		cs.MaxMS = ms
	}
	mem.ByCommand[command] = cs

	b, err := json.MarshalIndent(mem, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func readRuntimeTime(path string) (*runtimeTimeMemory, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var mem runtimeTimeMemory
	if err := json.Unmarshal(b, &mem); err != nil {
		return nil, err
	}
	if mem.ByCommand == nil {
		mem.ByCommand = make(map[string]runtimeCommand)
	}
	return &mem, nil
}
