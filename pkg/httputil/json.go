package httputil

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// WriteJSON serializes payload as JSON and writes it to the response.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// HandleHealth returns a standard /healthz handler with uptime info.
func HandleHealth(serviceName string) http.HandlerFunc {
	startedAt := time.Now()
	return func(w http.ResponseWriter, _ *http.Request) {
		uptime := time.Since(startedAt).Truncate(time.Second)
		body := map[string]string{
			"status":     "ok",
			"started_at": startedAt.UTC().Format(time.RFC3339),
			"uptime":     uptime.String(),
		}
		if serviceName != "" {
			body["service"] = serviceName
		}
		WriteJSON(w, http.StatusOK, body)
	}
}

// ParseLimit parses a pagination limit from a query param string.
func ParseLimit(raw string, defaultLimit, maxLimit int) int {
	if defaultLimit <= 0 {
		defaultLimit = 50
	}
	if maxLimit <= 0 {
		maxLimit = 200
	}
	s := strings.TrimSpace(raw)
	if s == "" {
		return defaultLimit
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return defaultLimit
	}
	if n > maxLimit {
		return maxLimit
	}
	return n
}

// CopyHeaders copies selected headers from src to dst.
func CopyHeaders(dst, src http.Header, keys ...string) {
	for _, key := range keys {
		if value := strings.TrimSpace(src.Get(key)); value != "" {
			dst.Set(key, value)
		}
	}
}

// CopyAllHeaders copies all headers from src to dst.
func CopyAllHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
