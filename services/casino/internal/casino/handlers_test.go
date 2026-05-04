package casino

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestInternalAuthMiddlewareRequiresConfiguredSecret(t *testing.T) {
	t.Setenv("CASINO_INTERNAL_SECRET", "")

	called := false
	handler := internalAuthMiddleware()(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if called {
		t.Fatal("next handler was called")
	}
}

func TestInternalAuthMiddlewareRejectsWrongSecret(t *testing.T) {
	t.Setenv("CASINO_INTERNAL_SECRET", "expected-secret")

	called := false
	handler := internalAuthMiddleware()(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	req.Header.Set("X-Internal-Secret", "wrong-secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if called {
		t.Fatal("next handler was called")
	}
}

func TestInternalAuthMiddlewareAcceptsConfiguredSecret(t *testing.T) {
	t.Setenv("CASINO_INTERNAL_SECRET", "expected-secret")

	handler := internalAuthMiddleware()(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	req.Header.Set("X-Internal-Secret", "expected-secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
}

func TestHandleConfigRejectsNonAdmin(t *testing.T) {
	store := NewStore(nil, NewEngine())
	hub := NewRouletteHub(store)
	handler := NewHandler(context.Background(), store, hub)
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

func TestExtractBonusInitData(t *testing.T) {
	t.Run("body wins", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/bonus/claim-subscription", strings.NewReader(`{"init_data":"body-value"}`))
		req.Header.Set("X-Telegram-Init-Data", "header-value")
		got, err := extractBonusInitData(req)
		if err != nil {
			t.Fatalf("extractBonusInitData() error = %v", err)
		}
		if got != "body-value" {
			t.Fatalf("got %q, want body-value", got)
		}
	})

	t.Run("header fallback", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/bonus/claim-subscription", nil)
		req.Header.Set("X-Init-Data", "header-value")
		got, err := extractBonusInitData(req)
		if err != nil {
			t.Fatalf("extractBonusInitData() error = %v", err)
		}
		if got != "header-value" {
			t.Fatalf("got %q, want header-value", got)
		}
	})
}

var _ = context.Background
