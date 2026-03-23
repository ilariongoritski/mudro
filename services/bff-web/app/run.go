package app

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/posts"
	"github.com/goritskimihail/mudro/internal/tgexport"
	"github.com/goritskimihail/mudro/services/bff-web/internal/bffweb"
)

func Run() {
	addr := envOr("BFF_WEB_ADDR", ":8086")
	dsn := config.DSN()
	movieCatalogURL := envOr("MOVIE_CATALOG_URL", "http://movie-catalog:8091")

	if err := config.ValidateRuntime("bff-web"); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	var tgVisiblePostIDs []string
	if ids, _, err := tgexport.LoadVisibleSourcePostIDsFromRepo(config.RepoRoot()); err == nil && len(ids) > 0 {
		tgVisiblePostIDs = ids
	}

	// Create reverse proxy for movie-catalog
	movieCatalogTarget, err := url.Parse(movieCatalogURL)
	if err != nil {
		log.Fatalf("invalid MOVIE_CATALOG_URL: %v", err)
	}
	movieCatalogProxy := httputil.NewSingleHostReverseProxy(movieCatalogTarget)

	// Create main handler
	mux := http.NewServeMux()

	// BFF endpoints
	bffHandler := bffweb.NewHandler(posts.NewService(pool, tgVisiblePostIDs), envOr("BFF_WEB_API_BASE_URL", config.APIBaseURL()))
	mux.Handle("/api/bff/web/v1/", bffHandler)

	// Movie catalog proxy
	mux.Handle("/api/movie-catalog/", http.StripPrefix("/api/movie-catalog", movieCatalogProxy))

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("bff-web listening on %s", addr)
		log.Printf("proxying /api/movie-catalog/* to %s", movieCatalogURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
