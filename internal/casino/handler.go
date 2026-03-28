package casino

import (
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

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	pool *pgxpool.Pool
	hub  *WSHub
}

func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{pool: pool, hub: NewWSHub()}
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

	var body struct {
		InitData string `json:"initData"`
	}
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

	var body struct {
		InitData string `json:"initData"`
	}
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

	acct, err := EnsureUserAccount(r.Context(), s.pool, auth.UserID, CasinoStartBalance())
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

	serverSeed := GenerateServerSeed()
	seedHash := HashServerSeed(serverSeed)

	round, err := PrepareRound(r.Context(), s.pool, auth.UserID, serverSeed, seedHash)
	if err != nil {
		log.Printf("prepare round: %v", err)
		writeJSON(w, 500, map[string]string{"error": "Internal error"})
		return
	}

	writeJSON(w, 200, map[string]any{
		"roundId":        round.ID,
		"serverSeedHash": seedHash,
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

	var body struct {
		RoundID        string  `json:"roundId"`
		BetAmount      float64 `json:"betAmount"`
		ClientSeed     string  `json:"clientSeed"`
		IdempotencyKey string  `json:"idempotencyKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, 400, map[string]string{"error": "Bad request"})
		return
	}

	result, err := PlaceBet(r.Context(), s.pool, BetInput{
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

	acct, err := GetUserAccount(r.Context(), s.pool, auth.UserID)
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

	var body struct {
		Amount float64 `json:"amount"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Amount <= 0 {
		body.Amount = CasinoFaucetAmount()
	}

	result, err := GrantFaucet(r.Context(), s.pool, auth.UserID, body.Amount)
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

	var userCount, roundCount int
	var houseBalance float64
	var totalBet, totalPayout float64

	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM casino_accounts WHERE type = 'user'`).Scan(&userCount)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM casino_rounds WHERE status = 'resolved'`).Scan(&roundCount)
	_ = s.pool.QueryRow(ctx, `SELECT COALESCE(balance, 0) FROM casino_accounts WHERE code = $1`, HouseAccountCode).Scan(&houseBalance)
	_ = s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(bet_amount), 0), COALESCE(SUM(payout_amount), 0) FROM casino_rounds WHERE status = 'resolved'`).Scan(&totalBet, &totalPayout)

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
		rows, err := s.pool.Query(ctx, `SELECT id, name, rtp, paytable, is_default FROM casino_rtp_profiles ORDER BY name`)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "query failed"})
			return
		}
		defer rows.Close()

		var profiles []map[string]any
		for rows.Next() {
			var id, name string
			var rtp float64
			var paytable json.RawMessage
			var isDefault bool
			if err := rows.Scan(&id, &name, &rtp, &paytable, &isDefault); err != nil {
				continue
			}
			profiles = append(profiles, map[string]any{
				"id": id, "name": name, "rtp": rtp, "paytable": json.RawMessage(paytable), "isDefault": isDefault,
			})
		}
		if profiles == nil {
			profiles = []map[string]any{}
		}
		writeJSON(w, 200, profiles)

	case http.MethodPost:
		var body struct {
			Name      string          `json:"name"`
			Rtp       float64         `json:"rtp"`
			Paytable  json.RawMessage `json:"paytable"`
			IsDefault bool            `json:"isDefault"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, 400, map[string]string{"error": "Bad request"})
			return
		}

		tiers, err := ParsePaytable(body.Paytable)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if err := ValidatePaytable(tiers, body.Rtp); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}

		if body.IsDefault {
			_, _ = s.pool.Exec(ctx, `UPDATE casino_rtp_profiles SET is_default = false WHERE is_default = true`)
		}

		var id string
		err = s.pool.QueryRow(ctx, `
			INSERT INTO casino_rtp_profiles (name, rtp, paytable, is_default)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (name) DO UPDATE SET rtp = $2, paytable = $3, is_default = $4, updated_at = now()
			RETURNING id
		`, body.Name, body.Rtp, body.Paytable, body.IsDefault).Scan(&id)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}

		ClearRtpCache("")
		adminKey := r.Header.Get("X-Admin-Key")
		log.Printf("[AUDIT] RTP profile upsert: id=%s name=%q rtp=%.2f isDefault=%v by=%s", id, body.Name, body.Rtp, body.IsDefault, maskKey(adminKey))
		writeJSON(w, 200, map[string]string{"id": id})

	default:
		// Handle DELETE via query param
		if r.Method == http.MethodDelete {
			id := strings.TrimPrefix(r.URL.Path, "/api/casino/admin/rtp/profiles/")
			if id == "" {
				writeJSON(w, 400, map[string]string{"error": "missing id"})
				return
			}
			_, _ = s.pool.Exec(ctx, `DELETE FROM casino_rtp_profiles WHERE id = $1`, id)
			ClearRtpCache("")
			adminKey := r.Header.Get("X-Admin-Key")
			log.Printf("[AUDIT] RTP profile deleted: id=%s by=%s", id, maskKey(adminKey))
		} else {
			writeJSON(w, 405, map[string]string{"error": "Method not allowed"})
		}
	}
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := s.pool.Query(ctx, `
		SELECT id, code, currency, balance FROM casino_accounts WHERE type = 'user' ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	var users []map[string]any
	for rows.Next() {
		var id, code, currency string
		var balance float64
		if err := rows.Scan(&id, &code, &currency, &balance); err != nil {
			continue
		}
		users = append(users, map[string]any{
			"id": id, "code": code, "currency": currency, "balance": balance,
		})
	}
	if users == nil {
		users = []map[string]any{}
	}
	writeJSON(w, 200, map[string]any{"items": users, "total": len(users)})
}

// ── Helpers ──

func writeJSON(w http.ResponseWriter, status int, data any) {
	httputil.WriteJSON(w, status, data)
}

func maskKey(key string) string {
	if len(key) <= 4 {
		return "***"
	}
	return key[:4] + "***"
}
