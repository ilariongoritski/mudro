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

func TestHandleConfigRejectsNonAdmin(t *testing.T) {
	handler := NewHandler(context.Background(), NewStore(nil, NewEngine()))
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
