package orchestrationapi

import (
	"io"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	baseURL    string
	httpClient *http.Client
}

func NewHandler(baseURL string) http.Handler {
	h := &Handler{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("/api/v1/orchestration/status", h.handleStatus)
	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if h.baseURL == "" {
		http.Error(w, "orchestration upstream is not configured", http.StatusServiceUnavailable)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, h.baseURL+"/api/orchestration/status", nil)
	if err != nil {
		http.Error(w, "failed to build upstream request", http.StatusInternalServerError)
		return
	}
	copySelectedHeaders(req.Header, r.Header, "Authorization", "X-Requested-With", "Cookie")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func copySelectedHeaders(dst, src http.Header, keys ...string) {
	for _, key := range keys {
		if value := strings.TrimSpace(src.Get(key)); value != "" {
			dst.Set(key, value)
		}
	}
}
