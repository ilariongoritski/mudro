package chat

import (
	"context"
	"encoding/json"
	"sync"
)

// Broadcaster is the interface satisfied by both Hub (in-memory) and RedisHub.
type Broadcaster interface {
	Start(ctx context.Context)
	Register(room string, client *Client)
	Unregister(room string, client *Client)
	Publish(msg Message)
	Close() error
}

type roomClient struct {
	room   string
	client *Client
}

type roomBroadcast struct {
	room    string
	payload []byte
}

type Hub struct {
	register   chan roomClient
	unregister chan roomClient
	broadcast  chan roomBroadcast
	stop       chan struct{}
	done       chan struct{}

	mu    sync.RWMutex
	rooms map[string]map[*Client]struct{}
	once  sync.Once
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan roomClient),
		unregister: make(chan roomClient),
		broadcast:  make(chan roomBroadcast, 32),
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
		rooms:      make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) Start(ctx context.Context) {
	go h.run(ctx)
}

func (h *Hub) Register(room string, client *Client) {
	select {
	case h.register <- roomClient{room: normalizeRoom(room), client: client}:
	case <-h.done:
		client.Close()
	}
}

func (h *Hub) Unregister(room string, client *Client) {
	select {
	case h.unregister <- roomClient{room: normalizeRoom(room), client: client}:
	case <-h.done:
		client.Close()
	}
}

func (h *Hub) Publish(msg Message) {
	payload, err := json.Marshal(socketEvent{
		Type:    "message",
		Message: &msg,
	})
	if err != nil {
		return
	}

	select {
	case h.broadcast <- roomBroadcast{room: normalizeRoom(msg.Room), payload: payload}:
	case <-h.done:
	}
}

func (h *Hub) Close() error {
	h.once.Do(func() {
		close(h.stop)
		<-h.done
	})
	return nil
}

func (h *Hub) run(ctx context.Context) {
	defer close(h.done)

	for {
		select {
		case <-ctx.Done():
			h.closeAll()
			return
		case <-h.stop:
			h.closeAll()
			return
		case item := <-h.register:
			h.mu.Lock()
			roomClients := h.rooms[item.room]
			if roomClients == nil {
				roomClients = make(map[*Client]struct{})
				h.rooms[item.room] = roomClients
			}
			roomClients[item.client] = struct{}{}
			h.mu.Unlock()
		case item := <-h.unregister:
			h.removeClient(item.room, item.client)
		case item := <-h.broadcast:
			h.mu.RLock()
			roomClients := h.rooms[item.room]
			for client := range roomClients {
				if ok := client.Enqueue(item.payload); !ok {
					go h.Unregister(item.room, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) removeClient(room string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	roomClients := h.rooms[room]
	if roomClients == nil {
		client.Close()
		return
	}
	if _, ok := roomClients[client]; ok {
		delete(roomClients, client)
		client.Close()
	}
	if len(roomClients) == 0 {
		delete(h.rooms, room)
	}
}

func (h *Hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for room, roomClients := range h.rooms {
		for client := range roomClients {
			client.Close()
		}
		delete(h.rooms, room)
	}
}
