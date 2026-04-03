package chat

import "time"

const DefaultRoom = "main"

type UserIdentity struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type Message struct {
	ID             int64        `json:"id"`
	ConversationID int64        `json:"conversation_id"`
	SenderID       int64        `json:"sender_id"`
	Room           string       `json:"room"` // Mapping for frontend backward compatibility
	Body           string       `json:"body"`
	CreatedAt      time.Time    `json:"created_at"`
	User           UserIdentity `json:"user"`
}

type listMessagesResponse struct {
	Items []Message `json:"items"`
}

type createMessageRequest struct {
	Room string `json:"room"`
	Body string `json:"body"`
}

type socketEvent struct {
	Type    string   `json:"type"`
	Message *Message `json:"message,omitempty"`
}
