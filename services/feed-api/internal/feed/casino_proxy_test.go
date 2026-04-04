package feed

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goritskimihail/mudro/internal/auth"
)

type casinoProxyUserRepo struct {
	user *auth.User
}

func (r *casinoProxyUserRepo) FindByLogin(context.Context, string) (*auth.User, error) {
	return nil, auth.ErrInvalidCredentials
}

func (r *casinoProxyUserRepo) FindByID(_ context.Context, id int64) (*auth.User, error) {
	if r.user != nil && r.user.ID == id {
		copy := *r.user
		return &copy, nil
	}
	return nil, auth.ErrNoSession
}

func (r *casinoProxyUserRepo) FindByTelegramID(context.Context, int64) (*auth.User, error) {
	return nil, auth.ErrNoSession
}

func (r *casinoProxyUserRepo) Create(context.Context, string, string, string) (*auth.User, error) {
	return nil, auth.ErrUserExists
}

func (r *casinoProxyUserRepo) CreateFromTelegram(context.Context, auth.TelegramUserParams) (*auth.User, error) {
	return nil, auth.ErrUserExists
}

func (r *casinoProxyUserRepo) UpdateTelegramName(context.Context, int64, string) error {
	return nil
}

func (r *casinoProxyUserRepo) HasActiveSubscription(context.Context, int64) (bool, error) {
	return false, nil
}

func (r *casinoProxyUserRepo) ListAll(context.Context) ([]auth.User, error) {
	return nil, nil
}

func (r *casinoProxyUserRepo) Count(context.Context) (int64, error) {
	return 0, nil
}

func (r *casinoProxyUserRepo) CountActiveSubscriptions(context.Context) (int64, error) {
	return 0, nil
}

func (r *casinoProxyUserRepo) AddSubscription(context.Context, int64, string, time.Duration) error {
	return nil
}

func TestHandleCasinoHistoryProxiesAuthenticatedUser(t *testing.T) {
	user := &auth.User{
		ID:       42,
		Username: "alice",
		Role:     "user",
	}
	email := "alice@example.com"
	user.Email = &email

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/history" {
			t.Fatalf("path = %s, want /history", r.URL.Path)
		}
		if r.URL.RawQuery != "limit=5" {
			t.Fatalf("raw query = %q, want limit=5", r.URL.RawQuery)
		}
		if got := r.Header.Get("X-User-ID"); got != "42" {
			t.Fatalf("X-User-ID = %q", got)
		}
		if got := r.Header.Get("X-User-Name"); got != "alice" {
			t.Fatalf("X-User-Name = %q", got)
		}
		if got := r.Header.Get("X-User-Email"); got != "alice@example.com" {
			t.Fatalf("X-User-Email = %q", got)
		}
		if got := r.Header.Get("X-User-Role"); got != "user" {
			t.Fatalf("X-User-Role = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"items":[]}`)
	}))
	defer upstream.Close()

	authSvc := auth.NewService(&casinoProxyUserRepo{user: user}, "test-secret")
	token, err := authSvc.IssueToken(user)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	s := NewServer(nil, nil, authSvc)
	s.casinoServiceURL = upstream.URL

	req := httptest.NewRequest(http.MethodGet, "/api/casino/history?limit=5", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(w.Body.String()); got != `{"items":[]}` {
		t.Fatalf("body = %q", got)
	}
}

func TestHandleCasinoBalanceRejectsMissingToken(t *testing.T) {
	s := NewServer(nil, nil, auth.NewService(&casinoProxyUserRepo{}, "test-secret"))
	s.casinoServiceURL = "http://casino.invalid"

	req := httptest.NewRequest(http.MethodGet, "/api/casino/balance", nil)
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleCasinoProfileProxiesAuthenticatedUser(t *testing.T) {
	user := &auth.User{
		ID:       7,
		Username: "bob",
		Role:     "user",
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/profile" {
			t.Fatalf("path = %s, want /profile", r.URL.Path)
		}
		if got := r.Header.Get("X-User-ID"); got != "7" {
			t.Fatalf("X-User-ID = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"user_id":7,"balance":500}`)
	}))
	defer upstream.Close()

	authSvc := auth.NewService(&casinoProxyUserRepo{user: user}, "test-secret")
	token, err := authSvc.IssueToken(user)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	s := NewServer(nil, nil, authSvc)
	s.casinoServiceURL = upstream.URL

	req := httptest.NewRequest(http.MethodGet, "/api/casino/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(w.Body.String()); got != `{"user_id":7,"balance":500}` {
		t.Fatalf("body = %q", got)
	}
}

func TestHandleCasinoBonusClaimSubscriptionProxiesInitData(t *testing.T) {
	user := &auth.User{
		ID:       9,
		Username: "carol",
		Role:     "user",
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bonus/claim-subscription" {
			t.Fatalf("path = %s, want /bonus/claim-subscription", r.URL.Path)
		}
		if got := r.Header.Get("X-Telegram-Init-Data"); got != "user=1&hash=abc" {
			t.Fatalf("X-Telegram-Init-Data = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"claimed","verification_status":"verified","state":{"user_id":9,"bonus_kind":"subscription","claimed":true}}`)
	}))
	defer upstream.Close()

	authSvc := auth.NewService(&casinoProxyUserRepo{user: user}, "test-secret")
	token, err := authSvc.IssueToken(user)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	s := NewServer(nil, nil, authSvc)
	s.casinoServiceURL = upstream.URL

	req := httptest.NewRequest(http.MethodPost, "/api/casino/bonus/claim-subscription", strings.NewReader(`{"init_data":"user=1&hash=abc"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-Init-Data", "user=1&hash=abc")
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(w.Body.String()); !strings.Contains(got, `"verification_status":"verified"`) {
		t.Fatalf("body = %q", got)
	}
}

func TestHandleCasinoBonusStateProxiesAuthenticatedUser(t *testing.T) {
	user := &auth.User{
		ID:       11,
		Username: "dave",
		Role:     "user",
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bonus/state" {
			t.Fatalf("path = %s, want /bonus/state", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"user_id":11,"bonus_kind":"subscription","claimed":false,"verification_status":"verification_required","free_spins_balance":0}`)
	}))
	defer upstream.Close()

	authSvc := auth.NewService(&casinoProxyUserRepo{user: user}, "test-secret")
	token, err := authSvc.IssueToken(user)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	s := NewServer(nil, nil, authSvc)
	s.casinoServiceURL = upstream.URL

	req := httptest.NewRequest(http.MethodGet, "/api/casino/bonus/state", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(w.Body.String()); !strings.Contains(got, `"verification_status":"verification_required"`) {
		t.Fatalf("body = %q", got)
	}
}

func TestHandleCasinoLiveFeedProxiesAuthenticatedUser(t *testing.T) {
	user := &auth.User{
		ID:       9,
		Username: "eve",
		Role:     "user",
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/live-feed" {
			t.Fatalf("path = %s, want /live-feed", r.URL.Path)
		}
		if got := r.Header.Get("X-User-ID"); got != "9" {
			t.Fatalf("X-User-ID = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"items":[]}`)
	}))
	defer upstream.Close()

	authSvc := auth.NewService(&casinoProxyUserRepo{user: user}, "test-secret")
	token, err := authSvc.IssueToken(user)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	s := NewServer(nil, nil, authSvc)
	s.casinoServiceURL = upstream.URL

	req := httptest.NewRequest(http.MethodGet, "/api/casino/live-feed", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(w.Body.String()); got != `{"items":[]}` {
		t.Fatalf("body = %q", got)
	}
}
