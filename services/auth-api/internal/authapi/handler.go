package authapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/goritskimihail/mudro/internal/api"
)

type Handler struct {
	authHandlers  *api.AuthHandlers
	adminHandlers *api.AdminHandlers
}

func NewHandler(authHandlers *api.AuthHandlers, adminHandlers *api.AdminHandlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealth)

	if authHandlers != nil {
		mux.HandleFunc("/api/v1/auth/register", authHandlers.HandleRegister)
		mux.HandleFunc("/api/v1/auth/login", authHandlers.HandleLogin)
		mux.HandleFunc("/api/v1/auth/telegram", authHandlers.HandleTelegramAuth)
		mux.HandleFunc("/api/v1/auth/logout", authHandlers.HandleLogout)
		mux.HandleFunc("/api/v1/auth/me", authHandlers.AuthMiddleware(authHandlers.HandleMe))
	} else {
		mux.HandleFunc("/api/v1/auth/register", serviceUnavailable)
		mux.HandleFunc("/api/v1/auth/login", serviceUnavailable)
		mux.HandleFunc("/api/v1/auth/telegram", serviceUnavailable)
		mux.HandleFunc("/api/v1/auth/logout", serviceUnavailable)
		mux.HandleFunc("/api/v1/auth/me", serviceUnavailable)
	}

	if authHandlers != nil && adminHandlers != nil {
		mux.HandleFunc("/api/v1/admin/users", authHandlers.AuthAdminMiddleware(adminHandlers.HandleGetUsers))
		mux.HandleFunc("/api/v1/admin/stats", authHandlers.AuthAdminMiddleware(adminHandlers.HandleGetStats))
	} else {
		mux.HandleFunc("/api/v1/admin/users", serviceUnavailable)
		mux.HandleFunc("/api/v1/admin/stats", serviceUnavailable)
	}

	return withCORS(mux)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func serviceUnavailable(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
