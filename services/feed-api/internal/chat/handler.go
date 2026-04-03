package chat

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/goritskimihail/mudro/internal/auth"
	"github.com/goritskimihail/mudro/internal/config"
)

var errUnauthorized = errors.New("unauthorized")

type Module struct {
	repo   *Repository
	hub    Broadcaster
	auth   *auth.Service
	cancel context.CancelFunc
	mux    *http.ServeMux
}

func NewModule(pool *pgxpool.Pool) (*Module, error) {
	ctx, cancel := context.WithCancel(context.Background())

	var hub Broadcaster
	if config.ChatHubBackend() == "redis" {
		rdb := redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr(),
			Password: config.RedisPassword(),
			DB:       config.RedisDB(),
		})
		slog.Info("chat: using Redis Pub/Sub hub", "addr", config.RedisAddr())
		hub = NewRedisHub(rdb)
	} else {
		hub = NewHub()
	}
	hub.Start(ctx)

	module := &Module{
		repo:   NewRepository(pool),
		hub:    hub,
		auth:   auth.NewService(auth.NewPgRepository(pool), config.JWTSecret()),
		cancel: cancel,
		mux:    http.NewServeMux(),
	}

	module.mux.HandleFunc("/api/chat/messages", module.handleMessages)
	module.mux.HandleFunc("/api/chat/ws", module.handleWS)

	return module, nil
}

func (m *Module) Handler() http.Handler {
	return m.mux
}

func (m *Module) Close() error {
	m.cancel()
	return m.hub.Close()
}

func (m *Module) handleMessages(w http.ResponseWriter, r *http.Request) {
	m.writeCORSHeaders(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// REST endpoints: only accept token from Authorization header (not query params).
	user, err := m.authenticateFromHeader(r)
	if err != nil {
		http.Error(w, errUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		m.handleListMessages(w, r, user)
	case http.MethodPost:
		m.handleCreateMessage(w, r, user)
	default:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (m *Module) handleListMessages(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	room := normalizeRoom(r.URL.Query().Get("room"))
	limit := parseLimit(r.URL.Query().Get("limit"), 50, 100)
	beforeID, err := parseOptionalInt64(r.URL.Query().Get("before_id"))
	if err != nil {
		http.Error(w, "invalid before_id", http.StatusBadRequest)
		return
	}

	items, err := m.repo.ListMessages(r.Context(), room, limit, beforeID)
	if err != nil {
		http.Error(w, "failed to load chat messages", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, listMessagesResponse{Items: items})
}

func (m *Module) handleCreateMessage(w http.ResponseWriter, r *http.Request, user *auth.User) {
	var req createMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	body := strings.TrimSpace(req.Body)
	if body == "" {
		http.Error(w, "message body is required", http.StatusBadRequest)
		return
	}
	if len(body) > 2000 {
		http.Error(w, "message body is too long", http.StatusBadRequest)
		return
	}

	msg, err := m.repo.InsertMessage(r.Context(), UserIdentity{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	}, req.Room, body)
	if err != nil {
		http.Error(w, "failed to store chat message", http.StatusInternalServerError)
		return
	}

	m.hub.Publish(msg)
	writeJSON(w, http.StatusCreated, msg)
}

func (m *Module) handleWS(w http.ResponseWriter, r *http.Request) {
	user, err := m.authenticateForWS(r)
	if err != nil {
		http.Error(w, errUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				return true
			}
			return isAllowedOrigin(origin)
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	room := normalizeRoom(r.URL.Query().Get("room"))
	client := NewClient(conn)
	m.hub.Register(room, client)

	readyPayload, _ := json.Marshal(socketEvent{
		Type: "ready",
		Message: &Message{
			Room: room,
			User: UserIdentity{
				ID:       user.ID,
				Username: user.Username,
				Role:     user.Role,
			},
			CreatedAt: time.Now().UTC(),
		},
	})
	if !client.Enqueue(readyPayload) {
		slog.Warn("chat: failed to enqueue ready payload", "user_id", user.ID, "room", room)
	}

	go client.WritePump(func() {
		m.hub.Unregister(room, client)
	})
	go client.ReadPump(func() {
		m.hub.Unregister(room, client)
	})
}

// authenticateFromHeader extracts the JWT exclusively from the Authorization header.
// Use this for REST endpoints where tokens must not appear in URLs.
func (m *Module) authenticateFromHeader(r *http.Request) (*auth.User, error) {
	tokenString := parseBearerToken(r.Header.Get("Authorization"))
	return m.authenticateToken(r.Context(), tokenString)
}

// authenticateForWS extracts the JWT from query params (required for WebSocket,
// since the browser WebSocket API does not support custom headers).
func (m *Module) authenticateForWS(r *http.Request) (*auth.User, error) {
	tokenString := strings.TrimSpace(r.URL.Query().Get("token"))
	if tokenString == "" {
		tokenString = parseBearerToken(r.Header.Get("Authorization"))
	}
	return m.authenticateToken(r.Context(), tokenString)
}

func (m *Module) authenticateToken(ctx context.Context, tokenString string) (*auth.User, error) {
	if tokenString == "" {
		return nil, errUnauthorized
	}

	claims, err := m.auth.ValidateToken(tokenString)
	if err != nil {
		return nil, errUnauthorized
	}

	userID, err := parseSubjectClaim(claims["sub"])
	if err != nil {
		return nil, errUnauthorized
	}

	user, err := m.auth.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errUnauthorized
	}
	return user, nil
}

func (m *Module) writeCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return
	}
	if !isAllowedOrigin(origin) {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Vary", "Origin")
}

func parseBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return ""
}

func parseSubjectClaim(raw any) (int64, error) {
	switch value := raw.(type) {
	case float64:
		return int64(value), nil
	case int64:
		return value, nil
	case string:
		return strconv.ParseInt(value, 10, 64)
	case json.Number:
		return value.Int64()
	default:
		return 0, jwt.ErrTokenMalformed
	}
}

func parseOptionalInt64(raw string) (*int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseLimit(raw string, fallback, max int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}

func normalizeRoom(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" {
		return DefaultRoom
	}

	var builder strings.Builder
	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_':
			builder.WriteRune(r)
		}
	}

	room := builder.String()
	if room == "" {
		return DefaultRoom
	}
	if len(room) > 32 {
		return room[:32]
	}
	return room
}

func isAllowedOrigin(origin string) bool {
	// Use explicitly configured origins when available.
	for _, allowed := range config.CORSAllowedOrigins() {
		if allowed == "*" || strings.EqualFold(strings.TrimSpace(allowed), origin) {
			return true
		}
	}

	// In dev mode, allow localhost variants.
	if config.MudroEnv() == "" || config.MudroEnv() == "dev" || config.MudroEnv() == "development" {
		parsedOrigin, err := url.Parse(origin)
		if err != nil {
			return false
		}
		host := parsedOrigin.Hostname()
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return true
		}
	}

	// Also allow API's own origin.
	if apiBase, err := url.Parse(config.APIBaseURL()); err == nil && apiBase.Host != "" {
		parsedOrigin, err := url.Parse(origin)
		if err == nil && parsedOrigin.Host == apiBase.Host {
			return true
		}
	}

	return false
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
