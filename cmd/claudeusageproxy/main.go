package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type usageEnvelope struct {
	Model string `json:"model"`
	Usage struct {
		InputTokens  int64 `json:"input_tokens"`
		OutputTokens int64 `json:"output_tokens"`
	} `json:"usage"`
	Error json.RawMessage `json:"error,omitempty"`
}

type usageEntry struct {
	Timestamp        string `json:"ts"`
	Model            string `json:"model"`
	Path             string `json:"path"`
	StatusCode       int    `json:"status_code"`
	InputTokens      int64  `json:"input_tokens"`
	OutputTokens     int64  `json:"output_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
	UpstreamBaseURL  string `json:"upstream_base_url"`
}

type usageSummary struct {
	UpdatedAt        string
	Requests         int64
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
}

type app struct {
	upstream  string
	usageLog  string
	tokenYAML string
	client    *http.Client
	mu        sync.Mutex
	summary   usageSummary
}

func main() {
	addr := envOr("MUDRO_CLAUDE_PROXY_ADDR", "127.0.0.1:8788")
	upstream := strings.TrimRight(envOr("MUDRO_CLAUDE_UPSTREAM_BASE_URL", "https://claude-api.filips-site.online"), "/")
	usageLog := envOr("MUDRO_CLAUDE_USAGE_LOG", filepath.Join(".skaro", "usage_log.jsonl"))
	tokenYAML := envOr("MUDRO_CLAUDE_TOKEN_USAGE", filepath.Join(".skaro", "token_usage.yaml"))

	srv := &app{
		upstream:  upstream,
		usageLog:  usageLog,
		tokenYAML: tokenYAML,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
	srv.summary = readSummary(tokenYAML)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", srv.handleHealth)
	mux.HandleFunc("/stats", srv.handleStats)
	mux.HandleFunc("/", srv.handleProxy)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 15 * time.Second,
	}

	log.Printf("claudeusageproxy listening on %s -> %s", addr, upstream)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func (a *app) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"upstream": a.upstream,
		"usageLog": a.usageLog,
		"summary":  a.tokenYAML,
	})
}

func (a *app) handleStats(w http.ResponseWriter, _ *http.Request) {
	a.mu.Lock()
	summary := a.summary
	a.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"updated_at":        summary.UpdatedAt,
		"requests":          summary.Requests,
		"prompt_tokens":     summary.PromptTokens,
		"completion_tokens": summary.CompletionTokens,
		"total_tokens":      summary.TotalTokens,
		"upstream":          a.upstream,
	})
}

func (a *app) handleProxy(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	targetURL := a.upstream + r.URL.Path
	if raw := r.URL.RawQuery; raw != "" {
		targetURL += "?" + raw
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	upstreamReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, bytes.NewReader(requestBody))
	if err != nil {
		http.Error(w, "failed to build upstream request", http.StatusInternalServerError)
		return
	}
	copyHeaders(upstreamReq.Header, r.Header)
	upstreamReq.Host = ""

	resp, err := a.client.Do(upstreamReq)
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read upstream response", http.StatusBadGateway)
		return
	}

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(responseBody)

	entry := usageEntry{
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		Path:            r.URL.Path,
		StatusCode:      resp.StatusCode,
		UpstreamBaseURL: a.upstream,
	}
	var envelope usageEnvelope
	if err := json.Unmarshal(responseBody, &envelope); err == nil {
		entry.Model = strings.TrimSpace(envelope.Model)
		entry.InputTokens = envelope.Usage.InputTokens
		entry.OutputTokens = envelope.Usage.OutputTokens
		entry.TotalTokens = entry.InputTokens + entry.OutputTokens
	}
	a.appendUsage(entry)
}

func (a *app) appendUsage(entry usageEntry) {
	if err := os.MkdirAll(filepath.Dir(a.usageLog), 0o755); err != nil {
		log.Printf("mkdir usage dir: %v", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(a.tokenYAML), 0o755); err != nil {
		log.Printf("mkdir summary dir: %v", err)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.OpenFile(a.usageLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("open usage log: %v", err)
		return
	}
	encoded, _ := json.Marshal(entry)
	_, _ = file.Write(append(encoded, '\n'))
	_ = file.Close()

	a.summary.Requests++
	a.summary.PromptTokens += entry.InputTokens
	a.summary.CompletionTokens += entry.OutputTokens
	a.summary.TotalTokens += entry.TotalTokens
	a.summary.UpdatedAt = entry.Timestamp

	if err := writeSummary(a.tokenYAML, a.summary); err != nil {
		log.Printf("write usage summary: %v", err)
	}
}

func readSummary(path string) usageSummary {
	data, err := os.ReadFile(path)
	if err != nil {
		return usageSummary{}
	}
	var summary usageSummary
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		switch key {
		case "updated_at":
			summary.UpdatedAt = value
		case "requests":
			summary.Requests = parseInt64(value)
		case "prompt_tokens":
			summary.PromptTokens = parseInt64(value)
		case "completion_tokens":
			summary.CompletionTokens = parseInt64(value)
		case "total_tokens":
			summary.TotalTokens = parseInt64(value)
		}
	}
	return summary
}

func writeSummary(path string, summary usageSummary) error {
	content := fmt.Sprintf(
		"updated_at: %q\nrequests: %d\nprompt_tokens: %d\ncompletion_tokens: %d\ntotal_tokens: %d\n",
		summary.UpdatedAt,
		summary.Requests,
		summary.PromptTokens,
		summary.CompletionTokens,
		summary.TotalTokens,
	)
	return os.WriteFile(path, []byte(content), 0o644)
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		switch strings.ToLower(key) {
		case "host", "content-length":
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func parseInt64(raw string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return n
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
