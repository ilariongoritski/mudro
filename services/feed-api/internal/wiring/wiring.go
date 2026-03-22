package wiring

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/internal/api"
	"github.com/goritskimihail/mudro/internal/auth"
	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/posts"
	"github.com/goritskimihail/mudro/internal/tgexport"
)

// NewHandler builds the legacy HTTP stack for feed-api and preserves current runtime behavior.
func NewHandler(pool *pgxpool.Pool) (http.Handler, error) {
	auth.SetSecret(config.JWTSecret())

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	authSvc := auth.NewService(pool, jwtSecret)

	var tgVisiblePostIDs []string
	if ids, path, err := tgexport.LoadVisibleSourcePostIDsFromRepo(config.RepoRoot()); err == nil && len(ids) > 0 {
		tgVisiblePostIDs = ids
		log.Printf("main: loaded telegram visibility filter (%d ids) from %s", len(ids), path)
	} else if err != nil {
		log.Printf("main: telegram visibility filter disabled: %v", err)
	}

	postsSvc := posts.NewService(pool, tgVisiblePostIDs)
	authHandlers := api.NewAuthHandlers(authSvc)
	adminHandlers := api.NewAdminHandlers(authSvc)

	return api.NewServer(pool, postsSvc, authHandlers, adminHandlers).Router(), nil
}
