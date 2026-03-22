package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/goritskimihail/mudro/internal/auth"
)

type contextKey string
const userContextKey = contextKey("user")

// UserFromContext extracts the user from context.
func UserFromContext(ctx context.Context) *auth.User {
	if u, ok := ctx.Value(userContextKey).(*auth.User); ok {
		return u
	}
	return nil
}

// AuthHandlers provides HTTP handlers for authentication.
type AuthHandlers struct {
	authSvc *auth.Service
}

// NewAuthHandlers creates a new AuthHandlers.
func NewAuthHandlers(svc *auth.Service) *AuthHandlers {
	return &AuthHandlers{authSvc: svc}
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type meResponse struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Email     *string `json:"email,omitempty"`
	Role      string  `json:"role"`
	IsPremium bool    `json:"is_premium"`
}

type tokenResponse struct {
	Token string     `json:"token"`
	User  meResponse `json:"user"`
}

// HandleRegister registers a new user with username and password.
func (h *AuthHandlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	username := strings.TrimSpace(req.Login)
	if username == "" || len(req.Password) < 6 {
		http.Error(w, "invalid username or password too short", http.StatusBadRequest)
		return
	}

	user, err := h.authSvc.Register(r.Context(), username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			http.Error(w, "username already taken", http.StatusConflict)
			return
		}
		log.Printf("auth register: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"user": meResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			IsPremium: user.IsPremium,
		},
	})
}

// HandleLogin authenticates a user and returns a JWT.
func (h *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, token, err := h.authSvc.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		log.Printf("auth login: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokenResponse{
		Token: token,
		User: meResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			IsPremium: user.IsPremium,
		},
	})
}

// HandleMe returns the currently authenticated user based on the JWT token.
func (h *AuthHandlers) HandleMe(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(meResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		IsPremium: user.IsPremium,
	})
}

// HandleLogout is a stub for JWT since tokens are stateless. Client should delete token.
func (h *AuthHandlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// AuthMiddleware wraps an http.HandlerFunc to require a valid JWT token.
func (h *AuthHandlers) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, "unauthorized - no token", http.StatusUnauthorized)
			return
		}

		claims, err := h.authSvc.ValidateToken(token)
		if err != nil {
			http.Error(w, "unauthorized - invalid token", http.StatusUnauthorized)
			return
		}

		sub, ok := claims["sub"].(float64)
		if !ok {
			http.Error(w, "unauthorized - invalid subject", http.StatusUnauthorized)
			return
		}

		user, err := h.authSvc.GetUserByID(r.Context(), int64(sub))
		if err != nil {
			http.Error(w, "unauthorized - user not found", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// AuthAdminMiddleware wraps an http.HandlerFunc to require a valid JWT token and 'admin' role.
func (h *AuthHandlers) AuthAdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return h.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil || user.Role != "admin" {
			http.Error(w, "forbidden - admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if c, err := r.Cookie("token"); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}
