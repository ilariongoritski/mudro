package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/goritskimihail/mudro/pkg/vercelapi"
)

var (
	once    sync.Once
	router  http.Handler
	initErr error
)

func initServer() {
	h, err := vercelapi.NewHandler()
	if err != nil {
		initErr = err
		return
	}
	router = h
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
