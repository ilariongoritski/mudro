package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/internal/api"
	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/ratelimit"
)

func main() {
	addr := config.APIAddr()
	dsn := config.DSN()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	baseHandler := api.NewServer(pool).Router()
	handler := withAPIRateLimit(baseHandler, config.APIRateLimitRPS(), config.APIRateLimitBurst())

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", addr)
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

func withAPIRateLimit(next http.Handler, rps, burst int) http.Handler {
	if rps <= 0 {
		return next
	}

	type clientEntry struct {
		lim *ratelimit.TokenBucket
	}
	var (
		mu      sync.Mutex
		clients = map[string]*clientEntry{}
		ttl     = 10 * time.Minute
	)

	cleanupTicker := time.NewTicker(1 * time.Minute)
	go func() {
		for range cleanupTicker.C {
			now := time.Now()
			mu.Lock()
			for ip, entry := range clients {
				if now.Sub(entry.lim.LastSeen()) > ttl {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		ip := clientIP(r)
		now := time.Now()
		mu.Lock()
		entry, ok := clients[ip]
		if !ok {
			entry = &clientEntry{lim: ratelimit.NewTokenBucket(rps, burst)}
			clients[ip] = entry
		}
		allowed := entry.lim.Allow(now)
		mu.Unlock()

		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if _, err := netip.ParseAddr(ip); err == nil {
				return ip
			}
		}
	}
	hostPort := strings.TrimSpace(r.RemoteAddr)
	if hostPort == "" {
		return "unknown"
	}
	if addr, err := netip.ParseAddrPort(hostPort); err == nil {
		return addr.Addr().String()
	}
	if addr, err := netip.ParseAddr(hostPort); err == nil {
		return addr.String()
	}
	return "unknown"
}
