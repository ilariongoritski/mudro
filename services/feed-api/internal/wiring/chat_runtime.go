package wiring

import (
	"context"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/services/feed-api/internal/chat"
)

func NewHandlerWithLifecycle(pool *pgxpool.Pool) (http.Handler, io.Closer, error) {
	baseHandler, err := NewHandler(context.Background(), pool)
	if err != nil {
		return nil, nil, err
	}

	chatModule, err := chat.NewModule(pool)
	if err != nil {
		return nil, nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/api/chat/", chatModule.Handler())
	mux.Handle("/api/chat", chatModule.Handler())
	mux.Handle("/", baseHandler)

	return mux, chatModule, nil
}
