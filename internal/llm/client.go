package llm

import "context"

// Message represents a single message in a chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the request sent to the LLM.
type ChatRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model,omitempty"`
}

// ChatResponse is the response from the LLM.
type ChatResponse struct {
	Content string `json:"content"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Client is the interface for LLM providers.
type Client interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
