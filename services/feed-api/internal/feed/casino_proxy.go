package feed

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	mhttputil "github.com/goritskimihail/mudro/pkg/httputil"
)

func (s *Server) handleCasinoBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/balance", false)
}

func (s *Server) handleCasinoHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/history", false)
}

func (s *Server) handleCasinoSpin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/spin", false)
}

func (s *Server) handleCasinoConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPut:
		s.proxyCasino(w, r, "/config", true)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) proxyCasino(w http.ResponseWriter, r *http.Request, upstreamPath string, adminOnly bool) {
	if s.casinoServiceURL == "" {
		http.Error(w, "casino service unavailable", http.StatusServiceUnavailable)
		return
	}

	user, err := s.authenticatedUserFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if adminOnly && user.Role != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var body io.Reader
	if r.Body != nil && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		body = bytes.NewReader(payload)
	}

	targetURL := s.casinoServiceURL + upstreamPath
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, body)
	if err != nil {
		http.Error(w, "failed to build upstream request", http.StatusInternalServerError)
		return
	}
	mhttputil.CopyHeaders(req.Header, r.Header, "Accept", "Content-Type")
	req.Header.Set("X-User-ID", strconv.FormatInt(user.ID, 10))
	req.Header.Set("X-User-Name", user.Username)
	if user.Email != nil {
		req.Header.Set("X-User-Email", *user.Email)
	}
	req.Header.Set("X-User-Role", user.Role)

	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "casino upstream unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	mhttputil.CopyAllHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
