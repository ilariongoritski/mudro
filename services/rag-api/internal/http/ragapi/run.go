package ragapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/rag"
)

const defaultAddr = ":8092"

func Run(ctx context.Context) error {
	client, err := rag.NewOpenAIClient(rag.OpenAIConfigFromEnv())
	if err != nil {
		return err
	}
	retriever, err := rag.NewQdrantRetrieverFromEnv()
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:              envOr("RAG_ADDR", defaultAddr),
		Handler:           NewHandler(rag.NewService(client, retriever, client)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      45 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		return server.Shutdown(closeCtx)
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func ValidateConfig() error {
	if strings.TrimSpace(os.Getenv("RAG_LLM_API_KEY")) == "" && strings.TrimSpace(os.Getenv("LLM_API_KEY")) == "" && strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")) == "" {
		return fmt.Errorf("RAG_LLM_API_KEY or LLM_API_KEY is required")
	}
	return nil
}
