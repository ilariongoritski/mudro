package chat

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4 * 1024
	sendBufferSize = 32
)

type Client struct {
	conn   *websocket.Conn
	hub    *Hub
	repo   *Repository
	roomID string
	user   Principal
	send   chan []byte
}

func NewClient(conn *websocket.Conn, hub *Hub, repo *Repository, roomID string, user Principal) *Client {
	return &Client{
		conn:   conn,
		hub:    hub,
		repo:   repo,
		roomID: normalizeRoomID(roomID),
		user:   user,
		send:   make(chan []byte, sendBufferSize),
	}
}

func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var incoming IncomingMessage
		if err := c.conn.ReadJSON(&incoming); err != nil {
			return
		}

		body := strings.TrimSpace(incoming.Body)
		if body == "" || utf8.RuneCountInString(body) > MaxBodyLength {
			continue
		}

		saved, err := c.repo.SaveMessage(ctx, Message{
			RoomID:   c.roomID,
			UserID:   c.user.ID,
			Username: c.user.Username,
			Body:     body,
		})
		if err != nil {
			continue
		}

		c.hub.Publish(saved)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			if _, err := writer.Write(message); err != nil {
				_ = writer.Close()
				return
			}

			if err := writer.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
