package chat

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestHubBroadcastsToRegisteredClient(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		roomID: DefaultRoomID,
		send:   make(chan []byte, 1),
	}

	hub.Register(client)
	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		hub.Publish(Message{
			ID:        42,
			RoomID:    DefaultRoomID,
			UserID:    "7",
			Username:  "tester",
			Body:      "hello",
			CreatedAt: time.Now(),
		})

		select {
		case frame := <-client.send:
			var event Event
			if err := json.Unmarshal(frame, &event); err != nil {
				t.Fatalf("unmarshal event: %v", err)
			}

			if event.Type != "chat:message" {
				t.Fatalf("expected chat:message, got %q", event.Type)
			}
			return
		case <-ticker.C:
		case <-deadline:
			t.Fatal("timed out waiting for broadcast")
		}
	}
}
