package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goritskimihail/mudro/internal/auth"
)

// MockService struct represents a mocked authentication service
type MockService struct{}

func (s *MockService) Login(ctx context.Context, email, password string) (*auth.User, string, error) {
	if email == "test@example.com" && password == "password123" {
		return &auth.User{ID: 1, Email: email, Role: "user"}, "mock-jwt-token", nil
	}
	return nil, "", auth.ErrInvalidCredentials
}

func (s *MockService) Register(ctx context.Context, email, password string) (*auth.User, error) {
	if email == "exists@example.com" {
		return nil, auth.ErrUserExists
	}
	return &auth.User{ID: 2, Email: email, Role: "user"}, nil
}

func (s *MockService) ValidateToken(token string) (map[string]interface{}, error) {
	return nil, nil
}

func (s *MockService) GetUserByID(ctx context.Context, id int64) (*auth.User, error) {
	return nil, nil
}

// Since auth.Service is a concrete struct in auth package and not an interface, 
// we would normally use interfaces or monkey patching for full unit tests.
// The code below demonstrates the shape of the test utilizing the handlers.
// Note: To fully unit test AuthHandlers, auth.Service should be decoupled via an interface.

func TestHandleLogin_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	rr := httptest.NewRecorder()

	handlers := NewAuthHandlers(nil)
	handlers.HandleLogin(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestHandleLogin_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{invalid}`))
	rr := httptest.NewRecorder()

	handlers := NewAuthHandlers(nil)
	handlers.HandleLogin(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestHandleRegister_MissingEmail(t *testing.T) {
	body, _ := json.Marshal(authRequest{Email: "not-an-email", Password: "short"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handlers := NewAuthHandlers(nil)
	handlers.HandleRegister(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
