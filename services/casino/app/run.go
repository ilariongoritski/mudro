package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goritskimihail/mudro/services/casino/internal/casino"
)

func Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := casino.OpenPool(ctx)
	if err != nil {
		log.Fatalf("casino open pool: %v", err)
	}
	defer pool.Close()

	store := casino.NewStore(pool, casino.NewEngine())
	if err := store.EnsureSeedConfig(ctx); err != nil {
		log.Fatalf("casino seed config: %v", err)
	}

	srv := &http.Server{
		Addr:              casino.Addr(),
		Handler:           casino.NewHandler(store).Router(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		log.Printf("casino listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("casino listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("casino shutdown: %v", err)
	}
}
