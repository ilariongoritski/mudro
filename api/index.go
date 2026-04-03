package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/goritskimihail/mudro/services/feed-api/vercelapi"
)

var (
	router  http.Handler
	once    sync.Once
	initErr error
)

func initServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	h, err := vercelapi.NewHandler(ctx)
	if err != nil {
		initErr = err
		return
	}
	router = h
}

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initServer)
	if initErr != nil {
		log.Printf("[VERCEL-API] init error: %v", initErr)
		http.Error(w, fmt.Sprintf("Backend Init Error: %v", initErr), http.StatusInternalServerError)
		return
	}
	router.ServeHTTP(w, r)
}
