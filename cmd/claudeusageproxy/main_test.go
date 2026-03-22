package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleProxyRejectsRemoteCallerWithoutToken(t *testing.T) {
	t.Parallel()

	var upstreamCalls int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"model":"claude-opus-4.6","usage":{"input_tokens":1,"output_tokens":2}}`))
	}))
	t.Cleanup(upstream.Close)

	srv := &app{
		upstream: upstream.URL,
		apiKey:   "sekret",
		client:   upstream.Client(),
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(`{"prompt":"hello"}`))
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()

	srv.handleProxy(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusForbidden)
	}
	if upstreamCalls != 0 {
		t.Fatalf("upstream should not have been called, got %d calls", upstreamCalls)
	}
}

func TestHandleProxyAllowsLoopbackAndInjectsServerKey(t *testing.T) {
	t.Parallel()

	var gotAuth string
	var gotApiKey string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotApiKey = r.Header.Get("X-Api-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"claude-opus-4.6","usage":{"input_tokens":3,"output_tokens":4}}`))
	}))
	t.Cleanup(upstream.Close)

	srv := &app{
		upstream: upstream.URL,
		apiKey:   "sekret",
		client:   upstream.Client(),
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(`{"prompt":"hello"}`))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleProxy(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}
	if gotAuth != "Bearer sekret" {
		t.Fatalf("unexpected auth header: got %q want %q", gotAuth, "Bearer sekret")
	}
	if gotApiKey != "sekret" {
		t.Fatalf("unexpected api key header: got %q want %q", gotApiKey, "sekret")
	}
}

func TestHandleProxyAllowsRemoteCallerWithProxyToken(t *testing.T) {
	t.Parallel()

	var gotAuth string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"claude-opus-4.6","usage":{"input_tokens":5,"output_tokens":6}}`))
	}))
	t.Cleanup(upstream.Close)

	srv := &app{
		upstream:   upstream.URL,
		apiKey:     "sekret",
		proxyToken: "proxy-secret",
		client:     upstream.Client(),
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(`{"prompt":"hello"}`))
	req.RemoteAddr = "198.51.100.7:1234"
	req.Header.Set("X-Mudro-Proxy-Token", "proxy-secret")
	rec := httptest.NewRecorder()

	srv.handleProxy(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}
	if gotAuth != "Bearer sekret" {
		t.Fatalf("unexpected auth header: got %q want %q", gotAuth, "Bearer sekret")
	}
}

func TestHandleStatsRejectsRemoteCallerWithoutToken(t *testing.T) {
	t.Parallel()

	srv := &app{}
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()

	srv.handleStats(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusForbidden)
	}
}

func TestHandleStatsAllowsLoopback(t *testing.T) {
	t.Parallel()

	srv := &app{
		summary: usageSummary{
			UpdatedAt:        "2026-03-22T19:29:10Z",
			Requests:         2,
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()

	srv.handleStats(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}
	if got := payload["total_tokens"]; got != float64(15) {
		t.Fatalf("unexpected total_tokens: got %v want %v", got, 15)
	}
}
