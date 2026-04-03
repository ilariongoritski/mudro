package casino

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		for _, allowed := range AllowedOrigins() {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

type WSHub struct {
	mu    sync.RWMutex
	conns map[string][]*websocket.Conn
}

func NewWSHub() *WSHub {
	return &WSHub{conns: make(map[string][]*websocket.Conn)}
}

func (h *WSHub) HandleUpgrade(w http.ResponseWriter, r *http.Request) {
	// Authenticate WebSocket connection via initData query param.
	initData := r.URL.Query().Get("initData")
	var userID string

	if initData != "" {
		auth, err := ValidateInitData(CasinoBotToken(), initData)
		if err != nil {
			http.Error(w, "unauthorized: invalid initData", http.StatusUnauthorized)
			return
		}
		userID = auth.UserID
	} else if CasinoDemoMode() {
		// Allow unauthenticated connections only in demo mode.
		userID = r.URL.Query().Get("userId")
		if userID == "" {
			userID = "tg_700001234"
		}
	} else {
		http.Error(w, "unauthorized: initData required", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}

	h.mu.Lock()
	h.conns[userID] = append(h.conns[userID], conn)
	h.mu.Unlock()

	if err := conn.WriteJSON(map[string]any{"type": "connection:ready", "userId": userID}); err != nil {
		slog.Debug("ws: send ready failed", "user_id", userID, "err", err)
	}

	// Read loop (keep alive, handle close)
	go func() {
		defer func() {
			h.remove(userID, conn)
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

func (h *WSHub) remove(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.conns[userID]
	for i, c := range conns {
		if c == conn {
			h.conns[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	if len(h.conns[userID]) == 0 {
		delete(h.conns, userID)
	}
}

func (h *WSHub) Emit(userID, eventType string, payload any) {
	data, err := json.Marshal(map[string]any{"type": eventType, "data": payload})
	if err != nil {
		return
	}

	h.mu.RLock()
	conns := h.conns[userID]
	h.mu.RUnlock()

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			slog.Debug("ws: broadcast write failed", "user_id", userID, "err", err)
		}
	}
}
