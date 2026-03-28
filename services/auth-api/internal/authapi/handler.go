package authapi

import (
	"net/http"

	"github.com/goritskimihail/mudro/pkg/httputil"
)

type Handler struct {
	authHandlers  *AuthHandlers
	adminHandlers *AdminHandlers
}

func NewHandler(authHandlers *AuthHandlers, adminHandlers *AdminHandlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", httputil.HandleHealth("auth-api"))

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

	return httputil.CORS(httputil.CORSConfig{
		SecurityHeaders: true,
	})(mux)
}

func serviceUnavailable(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
}
