# LLM test tool
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/llm"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	apiKey := config.LLMAPIKey()
	model := config.LLMModel()
	baseURL := config.LLMBaseURL()

	if apiKey == "" {
		fmt.Println("LLM_API_KEY not set")
		os.Exit(1)
	}

	client, err := llm.NewClient("codexsale", baseURL, apiKey, model)
	if err != nil {
		logger.Error("failed to create client", "err", err)
		os.Exit(1)
	}

	req := llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "Say hello in one sentence."},
		},
	}

	resp, err := client.Chat(context.Background(), req)
	if err != nil {
		logger.Error("LLM call failed", "err", err)
		os.Exit(1)
	}

	fmt.Printf("Model: %s\nResponse: %s\n", resp.Model, resp.Content)
}
