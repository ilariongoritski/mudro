package chat

import "time"

const (
	DefaultRoomID = "general"
	DefaultLimit  = 50
	MaxBodyLength = 2000
)

type Message struct {
	ID        int64     `json:"id"`
	RoomID    string    `json:"roomId"`
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type IncomingMessage struct {
	Body string `json:"body"`
}

type Principal struct {
	ID       string
	Username string
}
