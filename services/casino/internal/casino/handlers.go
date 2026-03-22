package casino

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("/balance", h.handleBalance)
	mux.HandleFunc("/history", h.handleHistory)
	mux.HandleFunc("/spin", h.handleSpin)
	mux.HandleFunc("/config", h.handleConfig)
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
	balance, err := h.store.GetBalance(r.Context(), actor)
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
		"balance":  balance,
		"rtp":      cfg.RTPPercent,
		"currency": "credits",
	})
}

func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
	actor, err := authContextFromHeaders(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
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
	var req struct {
		Bet int64 `json:"bet"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := h.store.Spin(r.Context(), actor, req.Bet)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInsufficientBalance) {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "max bet") {
			status = http.StatusBadRequest
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
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
