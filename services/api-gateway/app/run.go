package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goritskimihail/mudro/services/api-gateway/internal/gateway"
	pkgconfig "github.com/goritskimihail/mudro/pkg/config"
)

func Run() {
	addr := pkgconfig.EnvOr("GATEWAY_ADDR", ":8085")
	handler, err := gateway.NewHandler(gateway.Config{
		FeedAPIURL:          pkgconfig.EnvOr("GATEWAY_FEED_API_URL", "http://127.0.0.1:8080"),
		BFFWebURL:           pkgconfig.EnvOr("GATEWAY_BFF_WEB_URL", "http://127.0.0.1:8086"),
		AuthAPIURL:          pkgconfig.EnvOr("GATEWAY_AUTH_API_URL", "http://127.0.0.1:8087"),
		OrchestrationAPIURL: pkgconfig.EnvOr("GATEWAY_ORCHESTRATION_API_URL", "http://127.0.0.1:8088"),
	})
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("api-gateway listening on %s", addr)
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


