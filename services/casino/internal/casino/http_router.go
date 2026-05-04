package casino

import (
	"context"
	"net/http"
	"time"
)

type Handler struct {
	store       *Store
	hub         *RouletteHub
	userLimiter *UserRateLimiter
}

func NewHandler(ctx context.Context, store *Store, hub *RouletteHub) *Handler {
	return &Handler{
		store:       store,
		hub:         hub,
		userLimiter: NewUserRateLimiter(ctx, 10, time.Minute),
	}
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
