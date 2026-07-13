package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusCreated, map[string]string{"hello": "world"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["hello"] != "world" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestHandleHealth(t *testing.T) {
	t.Run("with service name", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		HandleHealth("bff-web").ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body["status"] != "ok" {
			t.Fatalf("expected status ok, got %q", body["status"])
		}
		if body["service"] != "bff-web" {
			t.Fatalf("expected service bff-web, got %q", body["service"])
		}
		if body["started_at"] == "" || body["uptime"] == "" {
			t.Fatalf("expected started_at and uptime to be set")
		}
	})

	t.Run("without service name", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		HandleHealth("").ServeHTTP(rec, req)

		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if _, ok := body["service"]; ok {
			t.Fatalf("expected no service key when name empty")
		}
	})
}

func TestParseLimit(t *testing.T) {
	cases := []struct {
		name         string
		raw          string
		defaultLimit int
		maxLimit     int
		want         int
	}{
		{"empty uses default", "", 20, 100, 20},
		{"invalid uses default", "abc", 20, 100, 20},
		{"negative uses default", "-5", 20, 100, 20},
		{"zero uses default", "0", 20, 100, 20},
		{"over max clamps", "999", 20, 100, 100},
		{"valid in range", "42", 20, 100, 42},
		{"whitespace trimmed", " 7 ", 20, 100, 7},
		{"default<=0 falls back to 50", "x", 0, 100, 50},
		{"max<=0 falls back to 200", "500", 20, 0, 200},
		{"valid with default<=0", "10", 0, 0, 10},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ParseLimit(c.raw, c.defaultLimit, c.maxLimit); got != c.want {
				t.Fatalf("ParseLimit(%q, %d, %d) = %d, want %d",
					c.raw, c.defaultLimit, c.maxLimit, got, c.want)
			}
		})
	}
}

func TestCopyHeaders(t *testing.T) {
	src := http.Header{}
	src.Set("Authorization", "Bearer x")
	src.Set("X-Request-Id", "abc")
	src.Set("X-Empty", "   ")

	dst := http.Header{}
	CopyHeaders(dst, src, "Authorization", "X-Request-Id", "X-Missing")

	if dst.Get("Authorization") != "Bearer x" {
		t.Fatalf("Authorization not copied")
	}
	if dst.Get("X-Request-Id") != "abc" {
		t.Fatalf("X-Request-Id not copied")
	}
	if _, ok := dst["X-Missing"]; ok {
		t.Fatalf("X-Missing should not be present")
	}
	if _, ok := dst["X-Empty"]; ok {
		t.Fatalf("X-Empty (whitespace) should be skipped")
	}
}

func TestCopyAllHeaders(t *testing.T) {
	src := http.Header{}
	src.Add("X-A", "1")
	src.Add("X-A", "2")
	src.Set("X-B", "3")

	dst := http.Header{}
	CopyAllHeaders(dst, src)

	if got := dst.Values("X-A"); len(got) != 2 || got[0] != "1" || got[1] != "2" {
		t.Fatalf("expected X-A to be copied with both values, got %v", got)
	}
	if dst.Get("X-B") != "3" {
		t.Fatalf("expected X-B=3")
	}
}

func TestCORS(t *testing.T) {
	t.Run("echoes allowed origin and security headers", func(t *testing.T) {
		cfg := CORSConfig{
			AllowedOrigins: []string{"https://example.com"},
			SecurityHeaders: true,
		}
		handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://example.com")
		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
			t.Fatalf("origin not echoed: %q", rec.Header().Get("Access-Control-Allow-Origin"))
		}
		if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Fatalf("expected nosniff security header")
		}
		if rec.Header().Get("X-Frame-Options") != "DENY" {
			t.Fatalf("expected X-Frame-Options DENY")
		}
	})

	t.Run("does not echo disallowed origin", func(t *testing.T) {
		cfg := CORSConfig{AllowedOrigins: []string{"https://example.com"}}
		handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://evil.com")
		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("disallowed origin should not be echoed")
		}
	})

	t.Run("OPTIONS returns 204", func(t *testing.T) {
		cfg := CORSConfig{AllowedOrigins: []string{"https://example.com"}}
		handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "https://example.com")
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204 for OPTIONS, got %d", rec.Code)
		}
	})

	t.Run("reads origins from env", func(t *testing.T) {
		t.Setenv("API_ALLOWED_ORIGINS", "https://env.com, https://env2.com")
		cfg := CORSConfig{}
		handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://env2.com")
		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "https://env2.com" {
			t.Fatalf("env origin not echoed: %q", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})
}
