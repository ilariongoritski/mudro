package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
)

type chatModeFile struct {
	UpdatedAt string          `json:"updated_at"`
	Chats     map[string]bool `json:"chats"`
}

var chatModeStore = struct {
	mu    sync.Mutex
	ready bool
	data  chatModeFile
}{
	data: chatModeFile{Chats: map[string]bool{}},
}

func (r *Runner) ChatModeStatus(chatID int64) ([]byte, error) {
	if err := r.loadChatModes(); err != nil {
		return nil, err
	}
	enabled := r.isChatModeEnabled(chatID)
	if enabled {
		if config.OpenRouterAPIKey() == "" {
			return []byte("Режим чата: ON\nOPENROUTER_API_KEY не задан, LLM-ответы недоступны.\nДобавь ключ в .env или используй команды."), nil
		}
		return []byte("Режим чата: ON\nТеперь можно писать обычные сообщения без команд."), nil
	}
	return []byte("Режим чата: OFF\nВключить: /chat on"), nil
}

func (r *Runner) SetChatMode(chatID int64, on bool) ([]byte, error) {
	if err := r.loadChatModes(); err != nil {
		return nil, err
	}
	chatModeStore.mu.Lock()
	defer chatModeStore.mu.Unlock()
	chatModeStore.data.Chats[strconv.FormatInt(chatID, 10)] = on
	chatModeStore.data.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := r.saveChatModesLocked(); err != nil {
		return nil, err
	}
	if on {
		return []byte("Режим чата включен. Пиши обычным текстом, бот будет отвечать как LLM-чат."), nil
	}
	return []byte("Режим чата выключен. Работают только команды."), nil
}

func (r *Runner) isChatModeEnabled(chatID int64) bool {
	chatModeStore.mu.Lock()
	defer chatModeStore.mu.Unlock()
	return chatModeStore.data.Chats[strconv.FormatInt(chatID, 10)]
}

func (r *Runner) loadChatModes() error {
	chatModeStore.mu.Lock()
	defer chatModeStore.mu.Unlock()
	if chatModeStore.ready {
		return nil
	}
	path := filepath.Join(r.RepoRoot, ".codex", "chat_mode.json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			chatModeStore.ready = true
			return nil
		}
		return fmt.Errorf("read chat_mode: %w", err)
	}
	var f chatModeFile
	if err := json.Unmarshal(b, &f); err != nil {
		return fmt.Errorf("parse chat_mode: %w", err)
	}
	if f.Chats == nil {
		f.Chats = map[string]bool{}
	}
	chatModeStore.data = f
	chatModeStore.ready = true
	return nil
}

func (r *Runner) saveChatModesLocked() error {
	path := filepath.Join(r.RepoRoot, ".codex", "chat_mode.json")
	b, err := json.MarshalIndent(chatModeStore.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal chat_mode: %w", err)
	}
	return os.WriteFile(path, b, 0o644)
}

func (r *Runner) handleChatText(updateText string) (string, bool) {
	t := strings.TrimSpace(updateText)
	if t == "" {
		return "", false
	}
	if strings.HasPrefix(t, "/") {
		return "", false
	}
	return t, true
}
