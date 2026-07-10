package llm

import (
	"fmt"

	"github.com/goritskimihail/mudro/internal/llm/providers/openai"
)

// NewClient creates an LLM client based on the provider.
func NewClient(provider, baseURL, apiKey, model string) (Client, error) {
	switch provider {
	case "codexsale", "openai", "zaprosu", "openrouter":
		return openai.New(baseURL, apiKey, model), nil
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s", provider)
	}
}
