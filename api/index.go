package handler

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	internalapi "github.com/goritskimihail/mudro/internal/api"
	"github.com/goritskimihail/mudro/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	once    sync.Once
	router  http.Handler
	initErr error
)

func initServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, config.DSN())
	if err != nil {
		initErr = err
		return
	}
	router = internalapi.NewServer(pool).Router()
}

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initServer)
	if initErr != nil {
		log.Printf("vercel init error: %v", initErr)
		http.Error(w, "init error", http.StatusInternalServerError)
		return
	}
	router.ServeHTTP(w, r)
}
