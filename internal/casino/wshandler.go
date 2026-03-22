package casino

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHub struct {
	mu    sync.RWMutex
	conns map[string][]*websocket.Conn
}

func NewWSHub() *WSHub {
	return &WSHub{conns: make(map[string][]*websocket.Conn)}
}

func (h *WSHub) HandleUpgrade(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "missing userId", http.StatusBadRequest)
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

	// Send ready
	_ = conn.WriteJSON(map[string]any{"type": "connection:ready", "userId": userID})

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
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
}
