package chat

import (
	"context"
	"slices"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListMessages(ctx context.Context, room string, limit int, beforeID *int64) ([]Message, error) {
	room = normalizeRoom(room)
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var before any
	if beforeID != nil {
		before = *beforeID
	}

	rows, err := r.pool.Query(ctx, `
		select id, room_name, user_id, username, user_role, body, created_at
		from chat_messages
		where room_name = $1
		  and ($2::bigint is null or id < $2)
		order by id desc
		limit $3
	`, room, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]Message, 0, limit)
	for rows.Next() {
		var msg Message
		if err := rows.Scan(
			&msg.ID,
			&msg.Room,
			&msg.User.ID,
			&msg.User.Username,
			&msg.User.Role,
			&msg.Body,
			&msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.Reverse(messages)
	return messages, nil
}

func (r *Repository) InsertMessage(ctx context.Context, user UserIdentity, room, body string) (Message, error) {
	room = normalizeRoom(room)

	var msg Message
	err := r.pool.QueryRow(ctx, `
		insert into chat_messages (room_name, user_id, username, user_role, body)
		values ($1, $2, $3, $4, $5)
		returning id, room_name, user_id, username, user_role, body, created_at
	`, room, user.ID, user.Username, user.Role, body).Scan(
		&msg.ID,
		&msg.Room,
		&msg.User.ID,
		&msg.User.Username,
		&msg.User.Role,
		&msg.Body,
		&msg.CreatedAt,
	)
	return msg, err
}
