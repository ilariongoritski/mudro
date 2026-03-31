package chat

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveMessage(ctx context.Context, msg Message) (Message, error) {
	row := r.pool.QueryRow(
		ctx,
		`
		INSERT INTO chat_messages (room_id, user_id, username, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, room_id, user_id, username, body, created_at
		`,
		normalizeRoomID(msg.RoomID),
		msg.UserID,
		msg.Username,
		strings.TrimSpace(msg.Body),
	)

	var saved Message
	err := row.Scan(
		&saved.ID,
		&saved.RoomID,
		&saved.UserID,
		&saved.Username,
		&saved.Body,
		&saved.CreatedAt,
	)
	return saved, err
}

func (r *Repository) LoadRecent(ctx context.Context, roomID string, limit int) ([]Message, error) {
	rows, err := r.pool.Query(
		ctx,
		`
		SELECT id, room_id, user_id, username, body, created_at
		FROM chat_messages
		WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		`,
		normalizeRoomID(roomID),
		clampLimit(limit),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]Message, 0, clampLimit(limit))
	for rows.Next() {
		var message Message
		if err := rows.Scan(
			&message.ID,
			&message.RoomID,
			&message.UserID,
			&message.Username,
			&message.Body,
			&message.CreatedAt,
		); err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}

	return messages, nil
}

func normalizeRoomID(roomID string) string {
	trimmed := strings.TrimSpace(roomID)
	if trimmed == "" {
		return DefaultRoomID
	}

	return trimmed
}

func clampLimit(limit int) int {
	if limit <= 0 || limit > DefaultLimit {
		return DefaultLimit
	}

	return limit
}
