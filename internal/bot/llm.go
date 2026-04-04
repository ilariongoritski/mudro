package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
)

func (r *Runner) AskMudro(query string) ([]byte, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []byte("Использование: /mudro <ваш вопрос>\nПример: /mudro что изменилось в боте за сегодня?"), nil
	}

	apiKey := config.OpenRouterAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is not set")
	}

	systemPrompt := strings.Join([]string{
		"Ты Mudro Assistant, технический помощник по репозиторию mudro.",
		"Отвечай по-русски, кратко и по делу.",
		"Приоритет: локальные факты проекта, затем общие рекомендации.",
		"Если данных не хватает, прямо скажи что именно нужно уточнить.",
	}, "\n")

	projectContext := r.buildProjectContext()

	body := chatCompletionsRequest{
		Model: config.OpenRouterModel(),
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Контекст проекта:\n" + projectContext + "\n\nЗапрос пользователя:\n" + query},
		},
		Temperature: 0.2,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal llm request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, config.OpenRouterBaseURL()+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build llm request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", config.APIBaseURL())
	req.Header.Set("X-Title", "Mudro")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	var out chatCompletionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode llm response: %w", err)
	}

	if resp.StatusCode >= 300 {
		msg := strings.TrimSpace(out.Error.Message)
		if msg == "" {
			msg = resp.Status
		}
		return nil, fmt.Errorf("llm api error: %s", msg)
	}

	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("llm api returned empty choices")
	}

	answer := strings.TrimSpace(out.Choices[0].Message.Content)
	if answer == "" {
		return nil, fmt.Errorf("llm api returned empty content")
	}
	return []byte(answer), nil
}

func (r *Runner) buildProjectContext() string {
	parts := []string{
		"repo: mudro",
		"root: " + r.RepoRoot,
		"allowed username: " + config.TelegramAllowedUsername(),
	}

	for _, item := range []struct {
		title string
		path  string
		limit int
	}{
		{title: "README", path: filepath.Join(r.RepoRoot, "README.md"), limit: 2500},
		{title: "TODO", path: filepath.Join(r.RepoRoot, ".codex", "todo.md"), limit: 1500},
		{title: "DONE", path: filepath.Join(r.RepoRoot, ".codex", "done.md"), limit: 1500},
	} {
		b, err := os.ReadFile(item.path)
		if err != nil {
			continue
		}
		text := strings.TrimSpace(string(b))
		if len(text) > item.limit {
			text = text[:item.limit] + "\n...(truncated)"
		}
		parts = append(parts, item.title+":\n"+text)
	}

	return strings.Join(parts, "\n\n")
}

type chatCompletionsRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}
