package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/internal/auth"
	"github.com/goritskimihail/mudro/internal/config"
	pkgconfig "github.com/goritskimihail/mudro/pkg/config"
	"github.com/goritskimihail/mudro/pkg/logger"
	authapi "github.com/goritskimihail/mudro/services/auth-api/internal/authapi"
)

func Run() {
	logger.Init("auth-api")
	addr := pkgconfig.EnvOr("AUTH_API_ADDR", ":8087")
	dsn := config.DSN()
	if err := config.ValidateRuntime("auth-api", "JWT_SECRET"); err != nil {
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

	auth.SetSecret(config.JWTSecret())
	authSvc := auth.NewService(auth.NewPgRepository(pool), config.JWTSecret())
	authSvc.SetTokenExpiry(time.Duration(config.JWTExpiryHours()) * time.Hour)
	handler := authapi.NewHandler(authapi.NewAuthHandlers(authSvc), authapi.NewAdminHandlers(authSvc))

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		log.Printf("auth-api listening on %s", addr)
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
