package vercelapi

import (
	"context"
	"net/http"
	"time"

	internalapi "github.com/goritskimihail/mudro/internal/api"
	"github.com/goritskimihail/mudro/internal/config"
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

	return internalapi.NewServer(pool, nil).Router(), nil
}
