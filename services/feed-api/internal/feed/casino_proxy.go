package feed

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	httputil "github.com/goritskimihail/mudro/pkg/httputil"
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

func (s *Server) handleCasinoBonusState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/bonus/state", false)
}

func (s *Server) handleCasinoBonusClaimSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/bonus/claim-subscription", false)
}

func (s *Server) handleCasinoBonusHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/bonus/history", false)
}

func (s *Server) handleCasinoProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/profile", false)
}

func (s *Server) handleCasinoActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/activity", false)
}

func (s *Server) handleCasinoLiveFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/live-feed", false)
}

func (s *Server) handleCasinoTopWins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/top-wins", false)
}

func (s *Server) handleCasinoReactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPost:
		s.proxyCasino(w, r, "/reactions", false)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCasinoRouletteState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/roulette/state", false)
}

func (s *Server) handleCasinoRouletteBets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/roulette/bets", false)
}

func (s *Server) handleCasinoRouletteHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/roulette/history", false)
}

func (s *Server) handleCasinoRouletteStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/roulette/stream", false)
}

func (s *Server) handleCasinoPlinkoConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/plinko/config", false)
}

func (s *Server) handleCasinoPlinkoState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/plinko/state", false)
}

func (s *Server) handleCasinoPlinkoDrop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/plinko/drop", false)
}

func (s *Server) handleCasinoBlackjackState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/blackjack/state", false)
}

func (s *Server) handleCasinoBlackjackStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/blackjack/start", false)
}

func (s *Server) handleCasinoBlackjackAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.proxyCasino(w, r, "/blackjack/action", false)
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
		limited := http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
		payload, err := io.ReadAll(limited)
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
	httputil.CopyHeaders(req.Header, r.Header, "Accept", "Content-Type", "X-Telegram-Init-Data", "X-Init-Data")
	req.Header.Set("X-User-ID", strconv.FormatInt(user.ID, 10))
	req.Header.Set("X-User-Name", user.Username)
	if user.Email != nil {
		req.Header.Set("X-User-Email", *user.Email)
	}
	req.Header.Set("X-User-Role", user.Role)

	client := s.httpClient
	if upstreamPath == "/roulette/stream" {
		client = s.sseClient
	}
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "casino upstream unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	httputil.CopyAllHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	if upstreamPath == "/roulette/stream" {
		if flusher, ok := w.(http.Flusher); ok {
			buf := make([]byte, 2048)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					_, _ = w.Write(buf[:n])
					flusher.Flush()
				}
				if err != nil {
					break
				}
			}
			return
		}
	}

	_, _ = io.Copy(w, resp.Body)
}
