package casino

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Handler struct {
	store       *Store
	userLimiter *UserRateLimiter
}

func NewHandler(ctx context.Context, store *Store) *Handler {
	return &Handler{store: store, userLimiter: NewUserRateLimiter(ctx, 10, time.Minute)}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)

	internal := internalAuthMiddleware()

	mux.Handle("/balance", internal(h.handleBalance))
	mux.Handle("/history", internal(h.handleHistory))
	mux.Handle("/spin", internal(h.handleSpin))
	mux.Handle("/config", internal(h.handleConfig))
	mux.Handle("/bonus/state", internal(h.handleBonusState))
	mux.Handle("/bonus/claim-subscription", internal(h.handleBonusClaimSubscription))
	mux.Handle("/bonus/history", internal(h.handleBonusHistory))

	mux.Handle("/roulette/state", internal(h.handleRouletteState))
	mux.Handle("/roulette/bets", internal(h.handleRouletteBets))
	mux.Handle("/roulette/instant-spin", internal(h.handleRouletteInstantSpin))
	mux.Handle("/roulette/history", internal(h.handleRouletteHistory))
	mux.Handle("/roulette/stream", internal(h.handleRouletteStream))

	mux.Handle("/profile", internal(h.handleProfile))
	mux.Handle("/activity", internal(h.handleActivity))
	mux.Handle("/live-feed", internal(h.handleLiveFeed))
	mux.Handle("/top-wins", internal(h.handleTopWins))
	mux.Handle("/reactions", internal(h.handleReactions))

	mux.Handle("/plinko/config", internal(h.handlePlinkoConfig))
	mux.Handle("/plinko/state", internal(h.handlePlinkoState))
	mux.Handle("/plinko/drop", internal(h.handlePlinkoDrop))

	mux.Handle("/blackjack/state", internal(h.handleBlackjackState))
	mux.Handle("/blackjack/start", internal(h.handleBlackjackStart))
	mux.Handle("/blackjack/action", internal(h.handleBlackjackAction))

	mux.Handle("/fairness/rotate-server-seed", internal(h.handleRotateServerSeed))
	mux.Handle("/fairness/client-seed", internal(h.handleUpdateClientSeed))
	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Health(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleBalance(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	balance, freeSpins, bonusClaimed, err := h.store.GetBalanceDetails(r.Context(), actor)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	cfg, err := h.store.GetConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"balance":            balance,
		"free_spins_balance": freeSpins,
		"bonus_claimed":      bonusClaimed || freeSpins > 0,
		"rtp":                cfg.RTPPercent,
		"currency":           "credits",
	})
}

func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	limit := parseLimit(r, 20)
	items, err := h.store.GetHistory(r.Context(), actor.UserID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) handleSpin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	if h.rateLimited(actor) {
		writeError(w, http.StatusTooManyRequests, errors.New("rate limit exceeded"))
		return
	}
	var req struct {
		Bet int64 `json:"bet"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := h.store.Spin(r.Context(), actor, req.Bet)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) handleBonusState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	state, err := h.store.GetBonusState(r.Context(), actor, parseLimit(r, 10))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *Handler) handleBonusClaimSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}

	initData, err := extractBonusInitData(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	resp, claimErr := h.store.ClaimSubscriptionBonus(r.Context(), actor, initData)
	if claimErr != nil {
		status := bonusClaimErrorStatus(claimErr)
		if resp != nil {
			writeJSON(w, status, resp)
			return
		}
		writeError(w, status, claimErr)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleBonusHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	items, err := h.store.GetBonusHistory(r.Context(), actor.UserID, parseLimit(r, 10))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, BonusClaimList{Items: items})
}

func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	if actor.Role != "admin" {
		writeError(w, http.StatusForbidden, ErrUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		cfg, err := h.store.GetConfig(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	case http.MethodPut:
		var cfg Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if err := h.store.UpdateConfig(r.Context(), cfg); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleRouletteState(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	state, err := h.store.GetCurrentRouletteState(r.Context(), actor.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *Handler) handleRouletteBets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	if h.rateLimited(actor) {
		writeError(w, http.StatusTooManyRequests, errors.New("rate limit exceeded"))
		return
	}

	var req RoulettePlaceBetsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.RoundID == 0 {
		state, err := h.store.GetCurrentRouletteState(r.Context(), actor.UserID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		req.RoundID = state.Round.ID
	}

	resp, err := h.store.PlaceRouletteBets(r.Context(), actor, req.RoundID, req.Bets)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleRouletteInstantSpin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	if h.rateLimited(actor) {
		writeError(w, http.StatusTooManyRequests, errors.New("rate limit exceeded"))
		return
	}

	var req RouletteInstantSpinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := h.store.InstantRouletteSpin(r.Context(), actor, req.Bets)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleRouletteHistory(w http.ResponseWriter, r *http.Request) {
	if _, err := authContextFromHeaders(r); err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	items, err := h.store.GetRouletteHistory(r.Context(), parseLimit(r, 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) handleRouletteStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("streaming unsupported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	sendState := func() bool {
		state, err := h.store.GetCurrentRouletteState(r.Context(), actor.UserID)
		if err != nil {
			_ = writeSSE(w, "error", map[string]string{"error": err.Error()})
			flusher.Flush()
			return false
		}
		if err := writeSSE(w, "state", state); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	if !sendState() {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if !sendState() {
				return
			}
		}
	}
}

func (h *Handler) handleProfile(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	profile, err := h.store.GetProfile(r.Context(), actor, parseLimit(r, 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (h *Handler) handleActivity(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	items, err := h.store.GetActivity(r.Context(), actor.UserID, parseLimit(r, 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, ActivityList{Items: items})
}

func (h *Handler) handleLiveFeed(w http.ResponseWriter, r *http.Request) {
	if _, err := authContextFromHeaders(r); err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	items, err := h.store.GetLiveFeed(r.Context(), parseLimit(r, 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, LiveFeedResponse{Items: items})
}

func (h *Handler) handleTopWins(w http.ResponseWriter, r *http.Request) {
	if _, err := authContextFromHeaders(r); err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	items, err := h.store.GetTopWins(r.Context(), parseLimit(r, 10))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, LiveFeedResponse{Items: items})
}

func (h *Handler) handleReactions(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := h.store.GetReactions(r.Context(), parseLimit(r, 20))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, ReactionList{Items: items})
	case http.MethodPost:
		var req ReactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := h.store.AddReaction(r.Context(), actor, req)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handlePlinkoConfig(w http.ResponseWriter, r *http.Request) {
	if _, err := authContextFromHeaders(r); err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	writeJSON(w, http.StatusOK, h.store.GetPlinkoConfig())
}

func (h *Handler) handlePlinkoState(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	state, err := h.store.GetPlinkoState(r.Context(), actor)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *Handler) handlePlinkoDrop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	if h.rateLimited(actor) {
		writeError(w, http.StatusTooManyRequests, errors.New("rate limit exceeded"))
		return
	}
	var req PlinkoDropRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := h.store.DropPlinko(r.Context(), actor, req)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func authContextFromHeaders(r *http.Request) (ParticipantInput, error) {
	userIDRaw := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if userIDRaw == "" {
		return ParticipantInput{}, ErrUnauthorized
	}
	userID, err := strconv.ParseInt(userIDRaw, 10, 64)
	if err != nil || userID <= 0 {
		return ParticipantInput{}, ErrUnauthorized
	}
	return ParticipantInput{
		UserID:   userID,
		Username: strings.TrimSpace(r.Header.Get("X-User-Name")),
		Email:    strings.TrimSpace(r.Header.Get("X-User-Email")),
		Role:     strings.TrimSpace(r.Header.Get("X-User-Role")),
	}, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeDomainError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, ErrInsufficientBalance):
		status = http.StatusConflict
	case errors.Is(err, ErrRoundClosed):
		status = http.StatusConflict
	case errors.Is(err, ErrNoActiveRound):
		status = http.StatusNotFound
	case errors.Is(err, ErrBonusAlreadyClaimed):
		status = http.StatusConflict
	case errors.Is(err, ErrBonusVerificationRequired):
		status = http.StatusPreconditionRequired
	case errors.Is(err, ErrBonusVerificationNotConfigured):
		status = http.StatusServiceUnavailable
	case errors.Is(err, ErrBonusVerificationDenied):
		status = http.StatusForbidden
	case errors.Is(err, ErrBonusVerificationUnavailable):
		status = http.StatusServiceUnavailable
	case strings.Contains(err.Error(), "max bet"),
		strings.Contains(err.Error(), "positive"),
		strings.Contains(err.Error(), "unsupported"),
		strings.Contains(err.Error(), "invalid"),
		strings.Contains(err.Error(), "duplicate"):
		status = http.StatusBadRequest
	}
	writeError(w, status, err)
}

func writeSSE(w http.ResponseWriter, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	return nil
}

func parseLimit(r *http.Request, fallback int) int {
	limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
	if limit <= 0 {
		return fallback
	}
	return limit
}

func extractBonusInitData(r *http.Request) (string, error) {
	var req BonusClaimRequest
	if r.Body != nil {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			return "", err
		}
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &req); err != nil {
				return "", err
			}
		}
	}
	if initData := strings.TrimSpace(req.InitData); initData != "" {
		return initData, nil
	}
	if initData := strings.TrimSpace(req.TelegramInitData); initData != "" {
		return initData, nil
	}
	if initData := strings.TrimSpace(r.Header.Get("X-Telegram-Init-Data")); initData != "" {
		return initData, nil
	}
	if initData := strings.TrimSpace(r.Header.Get("X-Init-Data")); initData != "" {
		return initData, nil
	}
	return "", nil
}

func bonusClaimErrorStatus(err error) int {
	switch {
	case errors.Is(err, ErrBonusAlreadyClaimed):
		return http.StatusConflict
	case errors.Is(err, ErrBonusVerificationRequired):
		return http.StatusPreconditionRequired
	case errors.Is(err, ErrBonusVerificationNotConfigured):
		return http.StatusServiceUnavailable
	case errors.Is(err, ErrBonusVerificationDenied):
		return http.StatusForbidden
	case errors.Is(err, ErrBonusVerificationUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
func (h *Handler) handleBlackjackState(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	state, err := h.store.BlackjackGetState(r.Context(), actor.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if state == nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": "no_game"})
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *Handler) handleBlackjackStart(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	if h.rateLimited(actor) {
		writeError(w, http.StatusTooManyRequests, errors.New("rate limit exceeded"))
		return
	}
	var req BlackjackGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	state, err := h.store.BlackjackStart(r.Context(), actor, req.Bet)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *Handler) handleBlackjackAction(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	var req BlackjackActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	state, err := h.store.BlackjackAction(r.Context(), actor, req.Action)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

type UserRateLimiter struct {
	mu       sync.Mutex
	requests map[int64]*userBucket
	rate     int
	window   time.Duration
	stopCh   chan struct{}
}

type userBucket struct {
	count       int
	windowStart time.Time
}

func NewUserRateLimiter(ctx context.Context, rate int, window time.Duration) *UserRateLimiter {
	l := &UserRateLimiter{
		requests: make(map[int64]*userBucket),
		rate:     rate,
		window:   window,
		stopCh:   make(chan struct{}),
	}
	go l.cleanupLoop(ctx)
	return l
}

func (l *UserRateLimiter) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(l.window)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.cleanup()
		case <-l.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (l *UserRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	staleCutoff := now.Add(-2 * l.window)
	for userID, bucket := range l.requests {
		if bucket.windowStart.Before(staleCutoff) {
			delete(l.requests, userID)
		}
	}
}

func (l *UserRateLimiter) Stop() {
	close(l.stopCh)
}

func (l *UserRateLimiter) Allow(userID int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	bucket, exists := l.requests[userID]

	if !exists || now.Sub(bucket.windowStart) > l.window {
		l.requests[userID] = &userBucket{count: 1, windowStart: now}
		return true
	}

	if bucket.count >= l.rate {
		return false
	}

	bucket.count++
	return true
}

func (h *Handler) rateLimited(actor ParticipantInput) bool {
	if h.userLimiter == nil {
		return false
	}
	return !h.userLimiter.Allow(actor.UserID)
}

func internalAuthMiddleware() func(http.HandlerFunc) http.Handler {
	secret := InternalSecret()
	return func(next http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret != "" {
				if r.Header.Get("X-Internal-Secret") != secret {
					http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
					return
				}
			}
			next(w, r)
		})
	}
}

func (h *Handler) handleRotateServerSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	newHash, err := h.store.RotateServerSeed(r.Context(), actor.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"new_server_seed_hash": newHash})
}

func (h *Handler) handleUpdateClientSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	var req struct {
		Seed string `json:"seed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.store.UpdateClientSeed(r.Context(), actor.UserID, req.Seed); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
