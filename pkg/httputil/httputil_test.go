package httputil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) && !strings.Contains(w.Body.String(), `"status": "ok"`) {
		t.Errorf("expected JSON body, got %s", w.Body.String())
	}
}

func TestHandleHealth(t *testing.T) {
	handler := HandleHealth("test-service")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
	body := w.Body.String()
	if !strings.Contains(body, `"status":"ok"`) && !strings.Contains(body, `"status": "ok"`) {
		t.Errorf("expected status ok in body: %s", body)
	}
	if !strings.Contains(body, `"service":"test-service"`) && !strings.Contains(body, `"service": "test-service"`) {
		t.Errorf("expected service name in body: %s", body)
	}
	if !strings.Contains(body, `"started_at"`) {
		t.Errorf("expected started_at in body: %s", body)
	}
	if !strings.Contains(body, `"uptime"`) {
		t.Errorf("expected uptime in body: %s", body)
	}
}

func TestHandleHealth_NoServiceName(t *testing.T) {
	handler := HandleHealth("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler(w, r)

	body := w.Body.String()
	if strings.Contains(body, `"service"`) {
		t.Errorf("expected no service field when empty, got: %s", body)
	}
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		defaultLimit int
		maxLimit     int
		expected     int
	}{
		{"empty string", "", 50, 200, 50},
		{"whitespace", "  ", 50, 200, 50},
		{"valid within range", "75", 50, 200, 75},
		{"exceeds max", "500", 50, 200, 200},
		{"below min", "0", 50, 200, 50},
		{"negative", "-10", 50, 200, 50},
		{"invalid", "abc", 50, 200, 50},
		{"custom defaults", "10", 10, 100, 10},
		{"custom defaults exceeded", "200", 10, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLimit(tt.raw, tt.defaultLimit, tt.maxLimit)
			if result != tt.expected {
				t.Errorf("ParseLimit(%q, %d, %d) = %d, want %d", tt.raw, tt.defaultLimit, tt.maxLimit, result, tt.expected)
			}
		})
	}
}

func TestCopyHeaders(t *testing.T) {
	src := http.Header{
		"X-Custom-Header": []string{"value1"},
		"Content-Type":    []string{"application/json"},
		"Authorization":   []string{"Bearer token"},
		"Empty-Header":    []string{""},
	}
	dst := http.Header{}

	CopyHeaders(dst, src, "X-Custom-Header", "Content-Type", "Missing-Header", "Empty-Header")

	if dst.Get("X-Custom-Header") != "value1" {
		t.Errorf("expected X-Custom-Header=value1, got %s", dst.Get("X-Custom-Header"))
	}
	if dst.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %s", dst.Get("Content-Type"))
	}
	if dst.Get("Missing-Header") != "" {
		t.Errorf("expected Missing-Header to be empty, got %s", dst.Get("Missing-Header"))
	}
	if dst.Get("Empty-Header") != "" {
		t.Errorf("expected Empty-Header to not be copied (empty value), got %s", dst.Get("Empty-Header"))
	}
}

func TestCopyAllHeaders(t *testing.T) {
	src := http.Header{
		"Header1": []string{"value1", "value2"},
		"Header2": []string{"value3"},
	}
	dst := http.Header{}

	CopyAllHeaders(dst, src)

	if len(dst["Header1"]) != 2 || dst["Header1"][0] != "value1" || dst["Header1"][1] != "value2" {
		t.Errorf("expected Header1 with 2 values, got %v", dst["Header1"])
	}
	if len(dst["Header2"]) != 1 || dst["Header2"][0] != "value3" {
		t.Errorf("expected Header2 with 1 value, got %v", dst["Header2"])
	}
}

func TestCORS_AllowsMatchingOrigin(t *testing.T) {
	handler := CORS(CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowCredentials: true,
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://example.com")

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler(next).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected ACAO header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Allow-Credentials header, got %s", w.Header().Get("Access-Control-Allow-Credentials"))
	}
	if !nextCalled {
		t.Errorf("next handler not called")
	}
}

func TestCORS_RejectsNonMatchingOrigin(t *testing.T) {
	handler := CORS(CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://evil.com")

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler(next).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no ACAO header for non-matching origin, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if !nextCalled {
		t.Errorf("next handler not called")
	}
}

func TestCORS_HandlesPreflight(t *testing.T) {
	handler := CORS(CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowHeaders:   []string{"X-Custom"},
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	r.Header.Set("Access-Control-Request-Method", "POST")
	r.Header.Set("Access-Control-Request-Headers", "X-Custom")

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler(next).ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected ACAO on preflight")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("expected Allow-Methods header")
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Errorf("expected Allow-Headers header")
	}
	if nextCalled {
		t.Errorf("next handler should not be called for preflight")
	}
}

func TestCORS_SecurityHeaders(t *testing.T) {
	handler := CORS(CORSConfig{
		AllowedOrigins:   []string{"https://example.com"},
		SecurityHeaders:  true,
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://example.com")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler(next).ServeHTTP(w, r)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options=nosniff")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("expected X-Frame-Options=DENY")
	}
}

func TestGzip_CompressesResponse(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strings.Repeat("hello world ", 100)))
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding=gzip, got %s", w.Header().Get("Content-Encoding"))
	}
	if w.Body.Len() >= 1100 {
		t.Errorf("expected compressed response smaller than original, got %d bytes", w.Body.Len())
	}
}

func TestGzip_NoCompressionWhenNotSupported(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Accept-Encoding header

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Encoding") != "" {
		t.Errorf("expected no Content-Encoding, got %s", w.Header().Get("Content-Encoding"))
	}
	if w.Body.String() != "hello" {
		t.Errorf("expected body 'hello', got %s", w.Body.String())
	}
}

func TestGzip_EmptyAcceptEncoding(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Encoding", "")

	handler.ServeHTTP(w, r)

	if w.Header().Get("Content-Encoding") != "" {
		t.Errorf("expected no compression for empty Accept-Encoding")
	}
}