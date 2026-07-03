package wiring

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/internal/auth"
	"github.com/goritskimihail/mudro/internal/chat"
	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/posts"
	"github.com/goritskimihail/mudro/internal/tgexport"
	"github.com/goritskimihail/mudro/services/feed-api/internal/feed"
)

// NewHandler builds the legacy HTTP stack for feed-api and preserves current runtime behavior.
// ctx controls the lifetime of background goroutines (e.g. chat hub); cancel it on shutdown.
func NewHandler(ctx context.Context, pool *pgxpool.Pool) (http.Handler, error) {

	var tgVisiblePostIDs []string
	if ids, path, err := tgexport.LoadVisibleSourcePostIDsFromRepo(config.RepoRoot()); err == nil && len(ids) > 0 {
		tgVisiblePostIDs = ids
		slog.Info("loaded telegram visibility filter", "count", len(ids), "path", path)
	} else if err != nil {
		slog.Warn("telegram visibility filter disabled", "err", err)
	}

	postsSvc := posts.NewService(pool, tgVisiblePostIDs)
	authSvc := auth.NewService(auth.NewPgRepository(pool), config.JWTSecret())
	chatRepo := chat.NewRepository(pool)
	chatHub := chat.NewHub()
	go chatHub.Run(ctx)
	chatHandler := chat.NewHandler(chatRepo, chatHub, authSvc)

	return feed.NewServer(postsSvc, chatHandler, authSvc).Router(), nil
}
