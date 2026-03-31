package chat

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/goritskimihail/mudro/internal/auth"
	"github.com/goritskimihail/mudro/internal/config"
)

type Handler struct {
	repo     *Repository
	hub      *Hub
	auth     *auth.Service
	upgrader websocket.Upgrader
}

func NewHandler(repo *Repository, hub *Hub, authService *auth.Service) *Handler {
	return &Handler{
		repo: repo,
		hub:  hub,
		auth: authService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := strings.TrimSpace(r.Header.Get("Origin"))
				if origin == "" {
					return config.MudroEnv() == "dev"
				}

				for _, allowed := range config.CORSAllowedOrigins() {
					if allowed == "*" || strings.EqualFold(strings.TrimSpace(allowed), origin) {
						return true
					}
				}

				if config.MudroEnv() == "dev" {
					return strings.HasPrefix(origin, "http://localhost") ||
						strings.HasPrefix(origin, "http://127.0.0.1") ||
						strings.HasPrefix(origin, "https://localhost") ||
						strings.HasPrefix(origin, "https://127.0.0.1")
				}

				return false
			},
		},
	}
}

func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request) {
	user, err := h.authenticate(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade уже записал HTTP-ответ об ошибке клиенту — только логируем.
		log.Printf("chat: websocket upgrade failed for user %s: %v", user.Username, err)
		return
	}

	client := NewClient(
		conn,
		h.hub,
		h.repo,
		r.URL.Query().Get("room"),
		Principal{
			ID:       strconv.FormatInt(user.ID, 10),
			Username: user.Username,
		},
	)

	h.hub.Register(client)

	frame, err := json.Marshal(Event{
		Type: "connection:ready",
		Data: map[string]string{
			"roomId": normalizeRoomID(client.roomID),
		},
	})
	if err == nil {
		client.send <- frame
	}

	go client.writePump()
	client.readPump(r.Context())
}

func (h *Handler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limit := DefaultLimit
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil {
			limit = parsed
		}
	}

	messages, err := h.repo.LoadRecent(r.Context(), r.URL.Query().Get("room"), limit)
	if err != nil {
		http.Error(w, "failed to load chat history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(messages)
}

func (h *Handler) authenticate(r *http.Request) (*auth.User, error) {
	if h.auth == nil {
		return nil, errors.New("chat auth service is required")
	}

	token := extractToken(r)
	if token == "" && config.MudroEnv() == "dev" {
		return &auth.User{
			ID:       0,
			Username: "local-dev",
			Role:     "admin",
		}, nil
	}

	claims, err := h.auth.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	subject, err := parseSubject(claims["sub"])
	if err != nil {
		return nil, err
	}

	return h.auth.GetUserByID(r.Context(), subject)
}

func extractToken(r *http.Request) string {
	if token := strings.TrimSpace(r.URL.Query().Get("token")); token != "" {
		return token
	}

	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[7:])
	}

	return ""
}

func parseSubject(value any) (int64, error) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), nil
	case string:
		return strconv.ParseInt(typed, 10, 64)
	case json.Number:
		return typed.Int64()
	default:
		return 0, strconv.ErrSyntax
	}
}
