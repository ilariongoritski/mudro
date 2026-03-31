package chat

import (
	"context"
	"encoding/json"
)

type broadcastMessage struct {
	roomID  string
	message Message
}

type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMessage
	rooms      map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan broadcastMessage, 64),
		rooms:      make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for _, clients := range h.rooms {
				for client := range clients {
					close(client.send)
				}
			}
			return
		case client := <-h.register:
			if client == nil {
				continue
			}

			roomID := normalizeRoomID(client.roomID)
			if _, ok := h.rooms[roomID]; !ok {
				h.rooms[roomID] = make(map[*Client]struct{})
			}
			h.rooms[roomID][client] = struct{}{}
		case client := <-h.unregister:
			h.removeClient(client)
		case payload := <-h.broadcast:
			h.broadcastMessage(payload)
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Publish(message Message) {
	h.broadcast <- broadcastMessage{
		roomID:  normalizeRoomID(message.RoomID),
		message: message,
	}
}

func (h *Hub) removeClient(client *Client) {
	if client == nil {
		return
	}

	roomID := normalizeRoomID(client.roomID)
	clients, ok := h.rooms[roomID]
	if !ok {
		return
	}

	if _, exists := clients[client]; exists {
		delete(clients, client)
		close(client.send)
	}

	if len(clients) == 0 {
		delete(h.rooms, roomID)
	}
}

func (h *Hub) broadcastMessage(payload broadcastMessage) {
	clients, ok := h.rooms[payload.roomID]
	if !ok {
		return
	}

	frame, err := json.Marshal(Event{
		Type: "chat:message",
		Data: payload.message,
	})
	if err != nil {
		return
	}

	for client := range clients {
		select {
		case client.send <- frame:
		default:
			delete(clients, client)
			close(client.send)
		}
	}

	if len(clients) == 0 {
		delete(h.rooms, payload.roomID)
	}
}
