package feed

import (
	"context"
	"net/http"
	"strings"

	"github.com/goritskimihail/mudro/internal/auth"
)

type contextKey string

const userContextKey = contextKey("user")

// UserFromContext extracts the user claims from context.
func UserFromContext(ctx context.Context) *auth.Claims {
	if u, ok := ctx.Value(userContextKey).(*auth.Claims); ok {
		return u
	}
	return nil
}

// WithAuth validates the JWT and injects claims into context.
func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" || token == header {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		claims, err := auth.ParseToken(token)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next(w, r.WithContext(ctx))
	}
}
