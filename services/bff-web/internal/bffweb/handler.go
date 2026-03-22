package bffweb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/posts"
)

type TimelineLoader interface {
	LoadPosts(ctx context.Context, beforeTS *time.Time, beforeID *int64, page *int, limit int, source string, sortOrder posts.SortOrder, query string) ([]posts.Post, *posts.Cursor, error)
}

type Handler struct {
	timeline   TimelineLoader
	apiBase    string
	httpClient *http.Client
}

type timelineResponse struct {
	Total      int            `json:"total"`
	Items      []timelineItem `json:"items"`
	NextCursor *posts.Cursor  `json:"next_cursor,omitempty"`
}

type timelineItem struct {
	ID            int64     `json:"id"`
	Source        string    `json:"source"`
	SourcePostID  string    `json:"source_post_id"`
	PostedAt      time.Time `json:"posted_at"`
	Text          *string   `json:"text,omitempty"`
	LikesCount    int       `json:"likes_count"`
	CommentsCount *int      `json:"comments_count,omitempty"`
}

type casinoWidgetResponse struct {
	Balance int64    `json:"balance"`
	RTP     float64  `json:"rtp"`
	Actions []string `json:"actions"`
}

func NewHandler(timeline TimelineLoader, apiBaseURL string) http.Handler {
	h := &Handler{
		timeline: timeline,
		apiBase:  strings.TrimRight(strings.TrimSpace(apiBaseURL), "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("/api/bff/web/v1/timeline", h.handleTimeline)
	mux.HandleFunc("/api/bff/web/v1/orchestration/status", h.handleOrchestrationStatus)
	mux.HandleFunc("/api/bff/web/v1/casino/widget", h.handleCasinoWidget)
	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleTimeline(w http.ResponseWriter, r *http.Request) {
	if h.timeline == nil {
		http.Error(w, "timeline service unavailable", http.StatusServiceUnavailable)
		return
	}

	limit := parseLimit(r.URL.Query().Get("limit"))
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	items, next, err := h.timeline.LoadPosts(r.Context(), nil, nil, nil, limit, source, posts.SortDesc, query)
	if err != nil {
		http.Error(w, "failed to load timeline", http.StatusInternalServerError)
		return
	}

	out := make([]timelineItem, 0, len(items))
	for _, item := range items {
		out = append(out, timelineItem{
			ID:            item.ID,
			Source:        item.Source,
			SourcePostID:  item.SourcePostID,
			PostedAt:      item.PublishedAt,
			Text:          item.Text,
			LikesCount:    item.LikesCount,
			CommentsCount: item.CommentsCount,
		})
	}
	writeJSON(w, http.StatusOK, timelineResponse{
		Total:      len(out),
		Items:      out,
		NextCursor: next,
	})
}

func (h *Handler) handleOrchestrationStatus(w http.ResponseWriter, r *http.Request) {
	h.proxyGET(w, r, "/api/orchestration/status", func(status int, headers http.Header, body []byte) {
		copyHeaders(w.Header(), headers)
		w.WriteHeader(status)
		_, _ = w.Write(body)
	})
}

func (h *Handler) handleCasinoWidget(w http.ResponseWriter, r *http.Request) {
	h.proxyGET(w, r, "/api/casino/balance", func(status int, _ http.Header, body []byte) {
		if status != http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write(body)
			return
		}

		var upstream struct {
			Balance  int64   `json:"balance"`
			RTP      float64 `json:"rtp"`
			Currency string  `json:"currency"`
		}
		if err := json.Unmarshal(body, &upstream); err != nil {
			http.Error(w, "invalid casino upstream response", http.StatusBadGateway)
			return
		}

		writeJSON(w, http.StatusOK, casinoWidgetResponse{
			Balance: upstream.Balance,
			RTP:     upstream.RTP,
			Actions: []string{"open-casino", "open-miniapp", "spin"},
		})
	})
}

func (h *Handler) proxyGET(w http.ResponseWriter, r *http.Request, upstreamPath string, handle func(status int, headers http.Header, body []byte)) {
	if h.apiBase == "" {
		http.Error(w, "API base URL is not configured", http.StatusServiceUnavailable)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, h.apiBase+upstreamPath, nil)
	if err != nil {
		http.Error(w, "failed to build upstream request", http.StatusInternalServerError)
		return
	}
	copySelectedHeaders(req.Header, r.Header, "Authorization", "X-Requested-With")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read upstream response", http.StatusBadGateway)
		return
	}
	handle(resp.StatusCode, resp.Header, body)
}

func parseLimit(raw string) int {
	const (
		defaultLimit = 20
		maxLimit     = 100
	)
	if strings.TrimSpace(raw) == "" {
		return defaultLimit
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n <= 0 {
		return defaultLimit
	}
	if n > maxLimit {
		return maxLimit
	}
	return n
}

func copySelectedHeaders(dst, src http.Header, keys ...string) {
	for _, key := range keys {
		if value := strings.TrimSpace(src.Get(key)); value != "" {
			dst.Set(key, value)
		}
	}
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) String() string {
	return fmt.Sprintf("bff-web(apiBase=%s)", h.apiBase)
}
