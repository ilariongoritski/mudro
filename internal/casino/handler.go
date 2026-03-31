package casino

import (
	"github.com/goritskimihail/mudro/internal/casino/contracts"
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/pkg/httputil"

	"github.com/goritskimihail/mudro/internal/casino/domain"
	"github.com/goritskimihail/mudro/internal/casino/usecase"
)

type Server struct {
	service *usecase.Service
	hub  *WSHub
}

func NewServer(service *usecase.Service) *Server {
	return &Server{service: service, hub: NewWSHub()}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok", "service": "casino"})
	})

	// Dev auth
	mux.HandleFunc("/api/casino/dev/init-data", s.handleDevInitData)

	// Auth
	mux.HandleFunc("/api/casino/auth/telegram", s.handleAuth)

	// Game
	mux.HandleFunc("/api/casino/game/prepare", s.withAuth(s.handlePrepare))
	mux.HandleFunc("/api/casino/game/bet", s.withAuth(s.handleBet))

	// Wallet
	mux.HandleFunc("/api/casino/wallet/balance", s.withAuth(s.handleBalance))
	mux.HandleFunc("/api/casino/wallet/faucet", s.withAuth(s.handleFaucet))

	// WebSocket
	mux.HandleFunc("/api/casino/ws", s.hub.HandleUpgrade)

	// Admin
	mux.HandleFunc("/api/casino/admin/stats", s.withAdmin(s.handleAdminStats))
	mux.HandleFunc("/api/casino/admin/rtp/profiles", s.withAdmin(s.handleAdminRtpProfiles))
	mux.HandleFunc("/api/casino/admin/users", s.withAdmin(s.handleAdminUsers))

	return httputil.CORS(httputil.CORSConfig{
		EnvVar:           "CASINO_ALLOWED_ORIGINS",
		DefaultOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://localhost:8080", "http://127.0.0.1:8080"},
		AllowHeaders:     []string{"Content-Type", "X-Init-Data", "X-Admin-Key"},
		AllowCredentials: true,
	})(mux)
}

// ── Middleware ──

type contextKey string

const authContextKey contextKey = "auth"

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		initData := r.Header.Get("X-Init-Data")
		if initData == "" {
			if bodyInitData, err := extractInitDataFromBody(r); err == nil && bodyInitData != "" {
				initData = bodyInitData
			}
		}

		auth, err := ValidateInitData(CasinoBotToken(), initData)
		if err != nil {
			if CasinoDemoMode() {
				// Demo fallback
				auth = &TelegramAuth{
					UserID:     "tg_700001234",
					TelegramID: 700001234,
					Username:   "demo_player",
					FirstName:  "Demo",
				}
			} else {
				writeJSON(w, 401, map[string]string{"error": "Unauthorized", "message": err.Error()})
				return
			}
		}

		// Enforce auth_date TTL (1 hour)
		if auth.AuthDate > 0 && time.Since(time.Unix(auth.AuthDate, 0)) > 1*time.Hour {
			writeJSON(w, 401, map[string]string{"error": "Unauthorized", "message": "expired initData"})
			return
		}

		ctx := context.WithValue(r.Context(), authContextKey, auth)
		next(w, r.WithContext(ctx))
	}
}

func extractInitDataFromBody(r *http.Request) (string, error) {
	if r.Method != http.MethodPost || r.Body == nil {
		return "", nil
	}

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	r.Body = io.NopCloser(bytes.NewReader(rawBody))
	if len(bytes.TrimSpace(rawBody)) == 0 {
		return "", nil
	}

	var body contracts.AuthRequest
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return "", err
	}
	return body.InitData, nil
}

func getAuth(r *http.Request) *TelegramAuth {
	if v, ok := r.Context().Value(authContextKey).(*TelegramAuth); ok {
		return v
	}
	return nil
}

func (s *Server) withAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expected := strings.TrimSpace(CasinoAdminKey())
		if expected == "" {
			writeJSON(w, 503, map[string]string{"error": "Admin key is not configured"})
			return
		}
		key := r.Header.Get("X-Admin-Key")
		if subtle.ConstantTimeCompare([]byte(key), []byte(expected)) != 1 {
			writeJSON(w, 401, map[string]string{"error": "Unauthorized"})
			return
		}
		next(w, r)
	}
}


// ── Handlers ──

func (s *Server) handleDevInitData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}
	// Block in production-like environments regardless of CASINO_DEMO_MODE.
	env := strings.ToLower(strings.TrimSpace(os.Getenv("MUDRO_ENV")))
	switch env {
	case "prod", "production", "stage", "staging":
		writeJSON(w, 403, map[string]string{"error": "Forbidden", "message": "Dev endpoint disabled in " + env})
		return
	}
	if !CasinoDemoMode() {
		writeJSON(w, 403, map[string]string{"error": "Forbidden", "message": "Demo mode disabled"})
		return
	}

	initData := DevInitData(CasinoBotToken(), 700001234)
	writeJSON(w, 200, map[string]any{"mode": "dev", "initData": initData})
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}

	var body contracts.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, 400, map[string]string{"error": "Bad request"})
		return
	}

	auth, err := ValidateInitData(CasinoBotToken(), body.InitData)
	if err != nil {
		if CasinoDemoMode() {
			auth = &TelegramAuth{
				UserID: "tg_700001234", TelegramID: 700001234,
				Username: "demo_player", FirstName: "Demo",
			}
		} else {
			writeJSON(w, 401, map[string]string{"error": "Unauthorized", "message": err.Error()})
			return
		}
	}

	acct, err := s.service.EnsureUserAccount(r.Context(), auth.UserID, CasinoStartBalance())
	if err != nil {
		log.Printf("ensure account: %v", err)
		writeJSON(w, 500, map[string]string{"error": "Internal error"})
		return
	}

	writeJSON(w, 200, map[string]any{
		"user": map[string]any{
			"id":         auth.UserID,
			"telegramId": strconv.FormatInt(auth.TelegramID, 10),
			"username":   auth.Username,
			"firstName":  auth.FirstName,
			"lastName":   auth.LastName,
		},
		"wallet": map[string]any{
			"balance":  acct.Balance,
			"currency": acct.Currency,
		},
		"authDate": auth.AuthDate,
	})
}

func (s *Server) handlePrepare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}
	auth := getAuth(r)
	if auth == nil {
		writeJSON(w, 401, map[string]string{"error": "Unauthorized"})
		return
	}

	round, err := s.service.PrepareRound(r.Context(), auth.UserID)
	if err != nil {
		log.Printf("prepare round: %v", err)
		writeJSON(w, 500, map[string]string{"error": "Internal error"})
		return
	}

	writeJSON(w, 200, map[string]any{
		"roundId":        round.ID,
		"serverSeedHash": round.ServerSeedHash,
		"createdAt":      round.CreatedAt,
	})
}

func (s *Server) handleBet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}
	auth := getAuth(r)
	if auth == nil {
		writeJSON(w, 401, map[string]string{"error": "Unauthorized"})
		return
	}

	var body contracts.BetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, 400, map[string]string{"error": "Bad request"})
		return
	}

	result, err := s.service.PlaceBet(r.Context(), domain.BetInput{
		UserID:         auth.UserID,
		RoundID:        body.RoundID,
		BetAmount:      body.BetAmount,
		ClientSeed:     body.ClientSeed,
		IdempotencyKey: body.IdempotencyKey,
	})
	if err != nil {
		log.Printf("bet error: %v", err)
		writeJSON(w, 400, map[string]string{"error": "BetFailed", "message": err.Error()})
		return
	}

	// Emit WebSocket events
	s.hub.Emit(auth.UserID, "game:roundResolved", result)
	s.hub.Emit(auth.UserID, "wallet:updated", map[string]float64{"balance": result.Balance})

	writeJSON(w, 200, result)
}

func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	auth := getAuth(r)
	if auth == nil {
		writeJSON(w, 401, map[string]string{"error": "Unauthorized"})
		return
	}

	acct, err := s.service.GetBalance(r.Context(), auth.UserID)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "Account not found"})
		return
	}

	writeJSON(w, 200, map[string]any{"balance": acct.Balance, "currency": acct.Currency})
}

func (s *Server) handleFaucet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}
	auth := getAuth(r)
	if auth == nil {
		writeJSON(w, 401, map[string]string{"error": "Unauthorized"})
		return
	}

	var body contracts.FaucetRequest
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Amount <= 0 {
		body.Amount = CasinoFaucetAmount()
	}

	result, err := s.service.GrantFaucet(r.Context(), auth.UserID, body.Amount)
	if err != nil {
		log.Printf("faucet error: %v", err)
		writeJSON(w, 400, map[string]string{"error": "FaucetFailed", "message": err.Error()})
		return
	}

	s.hub.Emit(auth.UserID, "wallet:updated", map[string]float64{"balance": result.Balance})

	writeJSON(w, 200, map[string]any{
		"amount":   result.Amount,
		"balance":  result.Balance,
		"currency": "МДР",
	})
}

// ── Admin Handlers ──

func (s *Server) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userCount, roundCount, houseBalance, totalBet, totalPayout, _ := s.service.GetStats(ctx)

	actualRtp := 0.0
	if totalBet > 0 {
		actualRtp = (totalPayout / totalBet) * 100
	}

	writeJSON(w, 200, map[string]any{
		"userCount":    userCount,
		"roundCount":   roundCount,
		"houseBalance": houseBalance,
		"totalBet":     totalBet,
		"totalPayout":  totalPayout,
		"actualRtp":    actualRtp,
	})
}

func (s *Server) handleAdminRtpProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		profiles, err := s.service.GetRTPProfiles(ctx)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "query failed"})
			return
		}
		writeJSON(w, 200, profiles)
	case http.MethodPost:
		var body contracts.UpsertRTPProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, 400, map[string]string{"error": "Bad request"})
			return
		}
		id, err := s.service.UpsertRTPProfile(ctx, body.Name, body.Rtp, body.Paytable, body.IsDefault)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]string{"id": id})
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, 400, map[string]string{"error": "missing id"})
			return
		}
		if err := s.service.DeleteRTPProfile(ctx, id); err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]string{"status": "ok"})
	default:
		writeJSON(w, 405, map[string]string{"error": "Method not allowed"})
	}
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := s.service.GetUsers(ctx, 100)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "query failed"})
		return
	}
	writeJSON(w, 200, map[string]any{"items": users, "total": len(users)})
}

// ── Helpers ──

func writeJSON(w http.ResponseWriter, status int, data any) {
	httputil.WriteJSON(w, status, data)
}

