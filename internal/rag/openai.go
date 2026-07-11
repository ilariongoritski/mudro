package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type OpenAIConfig struct{ BaseURL, APIKey, Model, EmbeddingModel string }

func OpenAIConfigFromEnv() OpenAIConfig {
	baseURL := env("RAG_LLM_BASE_URL", env("LLM_BASE_URL", env("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1")))
	return OpenAIConfig{BaseURL: strings.TrimRight(baseURL, "/"), APIKey: env("RAG_LLM_API_KEY", env("LLM_API_KEY", env("OPENROUTER_API_KEY", ""))), Model: env("RAG_LLM_MODEL", env("LLM_MODEL", env("OPENROUTER_MODEL", "openai/gpt-4.1-mini"))), EmbeddingModel: env("RAG_EMBEDDING_MODEL", "openai/text-embedding-3-small")}
}

type OpenAIClient struct {
	config OpenAIConfig
	http   *http.Client
}

func NewOpenAIClient(config OpenAIConfig) (*OpenAIClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("RAG_LLM_API_KEY or LLM_API_KEY is required")
	}
	return &OpenAIClient{config: config, http: &http.Client{Timeout: 40 * time.Second}}, nil
}
func (c *OpenAIClient) Embed(ctx context.Context, text string) ([]float64, error) {
	var response struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := c.post(ctx, "/embeddings", map[string]any{"model": c.config.EmbeddingModel, "input": text}, &response); err != nil {
		return nil, err
	}
	if len(response.Data) == 0 || len(response.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("embedding API returned no vector: %s", response.Error.Message)
	}
	return response.Data[0].Embedding, nil
}
func (c *OpenAIClient) Generate(ctx context.Context, question string, sources []Source) (string, error) {
	var contextParts []string
	for _, source := range sources {
		contextParts = append(contextParts, fmt.Sprintf("[%s]\n%s", source.Path, source.Excerpt))
	}
	body := map[string]any{"model": c.config.Model, "temperature": 0.1, "messages": []map[string]string{{"role": "system", "content": "You answer only from supplied Mudro technical documentation. Answer in Russian, concise. If evidence is insufficient say so. Cite source paths in square brackets."}, {"role": "user", "content": "Documentation:\n" + strings.Join(contextParts, "\n\n") + "\n\nQuestion: " + question}}}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := c.post(ctx, "/chat/completions", body, &response); err != nil {
		return "", err
	}
	if len(response.Choices) == 0 || strings.TrimSpace(response.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("LLM returned no answer: %s", response.Error.Message)
	}
	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}
func (c *OpenAIClient) post(ctx context.Context, path string, payload any, out any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("LLM request failed: %s", resp.Status)
	}
	return nil
}
func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
