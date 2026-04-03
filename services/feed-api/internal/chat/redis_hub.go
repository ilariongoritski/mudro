package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
)

// RedisHub is a multi-instance Hub backed by Redis Pub/Sub.
// Messages published on any instance are fanned out to local clients
// via a per-room Redis channel "chat:{room}".
//
// Enable with CHAT_HUB_BACKEND=redis.
type RedisHub struct {
	rdb *redis.Client

	mu    sync.RWMutex
	rooms map[string]map[*Client]struct{} // room → local clients

	// internal message dispatch
	local chan roomBroadcast

	stop chan struct{}
	done chan struct{}
	once sync.Once
}

func NewRedisHub(rdb *redis.Client) *RedisHub {
	return &RedisHub{
		rdb:   rdb,
		rooms: make(map[string]map[*Client]struct{}),
		local: make(chan roomBroadcast, 64),
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
}

func (h *RedisHub) Start(ctx context.Context) {
	go h.dispatchLoop(ctx)
}

func (h *RedisHub) Register(room string, client *Client) {
	room = normalizeRoom(room)
	h.mu.Lock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]struct{})
		// Start a Redis subscriber for this room the first time a client joins.
		go h.subscribe(room)
	}
	h.rooms[room][client] = struct{}{}
	h.mu.Unlock()
}

func (h *RedisHub) Unregister(room string, client *Client) {
	room = normalizeRoom(room)
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

// Publish serialises msg and publishes it to the Redis channel for its room.
// All instances (including this one) receive it via subscribe().
func (h *RedisHub) Publish(msg Message) {
	payload, err := json.Marshal(socketEvent{Type: "message", Message: &msg})
	if err != nil {
		slog.Error("redis hub: marshal message", "err", err)
		return
	}
	ctx := context.Background()
	if err := h.rdb.Publish(ctx, redisChannel(msg.Room), payload).Err(); err != nil {
		slog.Error("redis hub: publish", "room", msg.Room, "err", err)
	}
}

func (h *RedisHub) Close() error {
	h.once.Do(func() {
		close(h.stop)
		<-h.done
	})
	return nil
}

// subscribe runs a blocking Redis PubSub subscription for one room.
// It exits when the hub is stopped or when no local clients remain.
func (h *RedisHub) subscribe(room string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ps := h.rdb.Subscribe(ctx, redisChannel(room))
	defer ps.Close()

	ch := ps.Channel()
	for {
		select {
		case <-h.stop:
			return
		case redisMsg, ok := <-ch:
			if !ok {
				return
			}
			// Check if any local clients remain; stop if empty.
			h.mu.RLock()
			count := len(h.rooms[room])
			h.mu.RUnlock()
			if count == 0 {
				return
			}

			select {
			case h.local <- roomBroadcast{room: room, payload: []byte(redisMsg.Payload)}:
			case <-h.stop:
				return
			}
		}
	}
}

// dispatchLoop fans-out locally-queued payloads to connected clients.
func (h *RedisHub) dispatchLoop(ctx context.Context) {
	defer close(h.done)
	for {
		select {
		case <-ctx.Done():
			h.closeAll()
			return
		case <-h.stop:
			h.closeAll()
			return
		case item := <-h.local:
			h.mu.RLock()
			for client := range h.rooms[item.room] {
				if !client.Enqueue(item.payload) {
					go h.Unregister(item.room, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *RedisHub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for room, clients := range h.rooms {
		for c := range clients {
			c.Close()
		}
		delete(h.rooms, room)
	}
}

func redisChannel(room string) string {
	return "chat:" + normalizeRoom(room)
}
