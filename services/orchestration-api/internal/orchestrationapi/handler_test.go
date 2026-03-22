package orchestrationapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	handler := NewHandler("http://example.invalid")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestStatusProxy(t *testing.T) {
	var gotAuth string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orchestration/status" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"branch":"main","commit":"abc123"}`))
	}))
	defer upstream.Close()

	handler := NewHandler(upstream.URL)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orchestration/status", nil)
	req.Header.Set("Authorization", "Bearer demo")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotAuth != "Bearer demo" {
		t.Fatalf("Authorization = %q, want Bearer demo", gotAuth)
	}
	if !strings.Contains(rec.Body.String(), `"branch":"main"`) {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
