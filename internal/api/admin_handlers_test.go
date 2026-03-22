package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminHandlers_ServiceUnavailable(t *testing.T) {
	t.Run("users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
		rr := httptest.NewRecorder()

		handlers := NewAdminHandlers(nil)
		handlers.HandleGetUsers(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
		}
	})

	t.Run("stats", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/stats", nil)
		rr := httptest.NewRecorder()

		handlers := NewAdminHandlers(nil)
		handlers.HandleGetStats(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
		}
	})
}
