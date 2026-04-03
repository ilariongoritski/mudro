package feed

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goritskimihail/mudro/internal/auth"
)

func TestHandleCasinoHistoryProxiesAuthenticatedUser(t *testing.T) {
	email := "admin@mudro.local"
	user := &auth.User{
		ID:       42,
		Username: "admin",
		Email:    &email,
		Role:     "admin",
	}

	authSvc := auth.NewService(casinoProxyUserRepo{user: user}, "test-secret-123456")
	token, err := authSvc.IssueToken(user)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/history" {
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "6" {
			t.Fatalf("unexpected forwarded query: %q", got)
		}
		if got := r.Header.Get("X-User-ID"); got != "42" {
			t.Fatalf("unexpected X-User-ID: %q", got)
		}
		if got := r.Header.Get("X-User-Name"); got != "admin" {
			t.Fatalf("unexpected X-User-Name: %q", got)
		}
		if got := r.Header.Get("X-User-Email"); got != email {
			t.Fatalf("unexpected X-User-Email: %q", got)
		}
		if got := r.Header.Get("X-User-Role"); got != "admin" {
			t.Fatalf("unexpected X-User-Role: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"items":[]}`)
	}))
	defer upstream.Close()

	server := &Server{
		authSvc:          authSvc,
		httpClient:       upstream.Client(),
		casinoServiceURL: upstream.URL,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/casino/history?limit=6", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	server.handleCasinoHistory(rec, req)

	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if got := rec.Body.String(); got != `{"items":[]}` {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestHandleCasinoBalanceRejectsMissingToken(t *testing.T) {
	server := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/api/casino/balance", nil)
	rec := httptest.NewRecorder()

	server.handleCasinoBalance(rec, req)

	resp := rec.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

type casinoProxyUserRepo struct {
	user *auth.User
}

func (r casinoProxyUserRepo) FindByLogin(context.Context, string) (*auth.User, error) {
	return nil, auth.ErrInvalidCredentials
}

func (r casinoProxyUserRepo) FindByID(_ context.Context, id int64) (*auth.User, error) {
	if r.user != nil && r.user.ID == id {
		copyUser := *r.user
		return &copyUser, nil
	}
	return nil, auth.ErrNoSession
}

func (r casinoProxyUserRepo) FindByTelegramID(context.Context, int64) (*auth.User, error) {
	return nil, auth.ErrNoSession
}

func (r casinoProxyUserRepo) Create(context.Context, string, string, string) (*auth.User, error) {
	return nil, auth.ErrUserExists
}

func (r casinoProxyUserRepo) CreateFromTelegram(context.Context, auth.TelegramUserParams) (*auth.User, error) {
	return nil, auth.ErrUserExists
}

func (r casinoProxyUserRepo) UpdateTelegramName(context.Context, int64, string) error {
	return nil
}

func (r casinoProxyUserRepo) HasActiveSubscription(context.Context, int64) (bool, error) {
	return false, nil
}

func (r casinoProxyUserRepo) ListAll(context.Context) ([]auth.User, error) {
	return nil, nil
}

func (r casinoProxyUserRepo) Count(context.Context) (int64, error) {
	return 0, nil
}

func (r casinoProxyUserRepo) CountActiveSubscriptions(context.Context) (int64, error) {
	return 0, nil
}

func (r casinoProxyUserRepo) AddSubscription(context.Context, int64, string, time.Duration) error {
	return nil
}
