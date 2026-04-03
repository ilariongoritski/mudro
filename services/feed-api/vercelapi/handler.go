package vercelapi

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/posts"
	"github.com/goritskimihail/mudro/services/feed-api/internal/feed"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewHandler initializes dependencies and returns the HTTP router for Vercel.
func NewHandler() (http.Handler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, config.DSN())
	if err != nil {
		return nil, err
	}

	// Safely log the database host to verify the connection on Vercel
	if config.DSN() != "" {
		log.Printf("init db: %s", strings.Split(config.DSN(), "@")[len(strings.Split(config.DSN(), "@"))-1])
	}

	postsSvc := posts.NewService(pool, nil)
	return feed.NewServer(postsSvc, nil, nil).Router(), nil
}
