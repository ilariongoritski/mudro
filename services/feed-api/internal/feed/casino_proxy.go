package feed

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/goritskimihail/mudro/internal/auth"
	mhttputil "github.com/goritskimihail/mudro/pkg/httputil"
)

func (s *Server) handleCasinoBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasinoRequest(w, r, "/balance")
}

func (s *Server) handleCasinoHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasinoRequest(w, r, "/history")
}

func (s *Server) handleCasinoSpin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasinoRequest(w, r, "/spin")
}

func (s *Server) handleCasinoConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPut:
		s.proxyCasinoRequest(w, r, "/config")
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) proxyCasinoRequest(w http.ResponseWriter, r *http.Request, upstreamPath string) {
	user, err := s.authenticateCasinoUser(r)
	if err != nil {
		writeCasinoProxyError(w, http.StatusUnauthorized, err.Error())
		return
	}

	upstreamURL := strings.TrimRight(s.casinoServiceURL, "/") + upstreamPath
	if r.URL.RawQuery != "" {
		upstreamURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, r.Body)
	if err != nil {
		writeCasinoProxyError(w, http.StatusInternalServerError, "failed to build casino request")
		return
	}

	copyCasinoHeader(req.Header, r.Header, "Content-Type", "Accept")
	req.Header.Set("X-User-ID", strconv.FormatInt(user.ID, 10))
	req.Header.Set("X-User-Name", user.Username)
	req.Header.Set("X-User-Role", user.Role)
	if user.Email != nil {
		req.Header.Set("X-User-Email", *user.Email)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		writeCasinoProxyError(w, http.StatusBadGateway, "casino service unavailable")
		return
	}
	defer resp.Body.Close()

	copyCasinoHeader(w.Header(), resp.Header, "Content-Type")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s *Server) authenticateCasinoUser(r *http.Request) (*auth.User, error) {
	if s.authSvc == nil {
		return nil, errors.New("auth service unavailable")
	}

	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return nil, errors.New("missing bearer token")
	}

	claims, err := s.authSvc.ValidateToken(strings.TrimSpace(token[7:]))
	if err != nil {
		return nil, errors.New("invalid token")
	}

	subject, ok := claims["sub"].(float64)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	user, err := s.authSvc.GetUserByID(r.Context(), int64(subject))
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func copyCasinoHeader(dst, src http.Header, keys ...string) {
	mhttputil.CopyHeaders(dst, src, keys...)
}

func writeCasinoProxyError(w http.ResponseWriter, status int, message string) {
	mhttputil.WriteJSON(w, status, map[string]string{"error": message})
}
