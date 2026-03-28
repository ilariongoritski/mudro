package httputil

import (
	"net/http"
	"os"
	"strings"
)

// CORSConfig defines CORS middleware behavior.
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins.
	// If empty, reads from the envVar or falls back to defaults.
	AllowedOrigins []string
	// EnvVar is the environment variable to read origins from (comma-separated).
	EnvVar string
	// DefaultOrigins used when AllowedOrigins is empty and env is not set.
	DefaultOrigins []string
	// AllowHeaders specifies allowed request headers.
	AllowHeaders []string
	// AllowCredentials allows credentials in CORS requests.
	AllowCredentials bool
	// SecurityHeaders adds X-Content-Type-Options and X-Frame-Options.
	SecurityHeaders bool
}

// CORS returns a middleware that applies CORS headers based on config.
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	origins := cfg.AllowedOrigins
	if len(origins) == 0 {
		envVar := cfg.EnvVar
		if envVar == "" {
			envVar = "API_ALLOWED_ORIGINS"
		}
		raw := strings.TrimSpace(os.Getenv(envVar))
		if raw != "" {
			for _, p := range strings.Split(raw, ",") {
				if s := strings.TrimSpace(p); s != "" {
					origins = append(origins, s)
				}
			}
		}
	}
	if len(origins) == 0 {
		origins = cfg.DefaultOrigins
	}
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173", "http://localhost:8080"}
	}

	allowHeaders := "Authorization, Content-Type, X-Requested-With"
	if len(cfg.AllowHeaders) > 0 {
		allowHeaders = strings.Join(cfg.AllowHeaders, ", ")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.SecurityHeaders {
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-Frame-Options", "DENY")
			}

			if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
				for _, a := range origins {
					if origin == a {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
						w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
						if cfg.AllowCredentials {
							w.Header().Set("Access-Control-Allow-Credentials", "true")
						}
						w.Header().Set("Vary", "Origin")
						break
					}
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
