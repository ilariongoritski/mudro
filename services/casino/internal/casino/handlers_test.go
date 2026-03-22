package casino

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStore struct{}

func TestAuthContextFromHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	req.Header.Set("X-User-ID", "42")
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-User-Name", "alice")
	req.Header.Set("X-User-Email", "alice@example.com")

	actor, err := authContextFromHeaders(req)
	if err != nil {
		t.Fatalf("authContextFromHeaders() error = %v", err)
	}
	if actor.UserID != 42 || actor.Role != "admin" || actor.Username != "alice" || actor.Email != "alice@example.com" {
		t.Fatalf("unexpected auth context: %#v", actor)
	}
}

func TestHandleConfigRejectsNonAdmin(t *testing.T) {
	handler := NewHandler(NewStore(nil, NewEngine()))
	req := httptest.NewRequest(http.MethodGet, "/config", strings.NewReader(""))
	req.Header.Set("X-User-ID", "7")
	rec := httptest.NewRecorder()

	handler.handleConfig(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusOK, map[string]string{"status": "ok"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "status") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

var _ = context.Background
