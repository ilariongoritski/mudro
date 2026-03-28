package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/internal/catalog/service"
	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/services/movie-catalog/internal/adapters/postgrescatalog"
	httpmoviecatalog "github.com/goritskimihail/mudro/services/movie-catalog/internal/http/moviecatalog"
)

const (
	defaultAddr = ":8091"
	defaultDSN  = "postgres://postgres:postgres@db:5432/movie_catalog?sslmode=disable"
)

func Run(ctx context.Context) error {
	addr := getenv("MOVIE_CATALOG_ADDR", defaultAddr)
	dsn := getenv("MOVIE_CATALOG_DB_DSN", defaultDSN)
	if err := config.ValidateRuntimeDSN("movie-catalog", dsn); err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	repository := postgrescatalog.NewRepository(pool)
	catalogService := service.NewCatalogService(repository)
	handler := httpmoviecatalog.NewHandler(catalogService, pool.Ping)

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	shutdownCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-shutdownCtx.Done():
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}

	closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(closeCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	return nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
