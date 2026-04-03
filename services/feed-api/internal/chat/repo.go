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

	convID, err := r.GetOrCreateConversation(ctx, room)
	if err != nil {
		return nil, err
	}

	var before any
	if beforeID != nil {
		before = *beforeID
	}

	rows, err := r.pool.Query(ctx, `
		select 
			m.id, 
			m.conversation_id, 
			m.sender_id, 
			m.body, 
			m.created_at,
			u.id, 
			u.login, 
			u.role,
			coalesce(u.avatar_url, '')
		from messages m
		left join users u on u.id = m.sender_id
		where m.conversation_id = $1
		  and ($2::bigint is null or m.id < $2)
		order by m.id desc
		limit $3
	`, convID, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]Message, 0, limit)
	for rows.Next() {
		var msg Message
		msg.Room = room
		if err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.Body,
			&msg.CreatedAt,
			&msg.User.ID,
			&msg.User.Username,
			&msg.User.Role,
			&msg.User.AvatarURL,
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
	convID, err := r.GetOrCreateConversation(ctx, room)
	if err != nil {
		return Message{}, err
	}

	var msg Message
	msg.Room = room
	err = r.pool.QueryRow(ctx, `
		insert into messages (conversation_id, sender_id, body)
		values ($1, $2, $3)
		returning id, conversation_id, sender_id, body, created_at
	`, convID, user.ID, body).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderID,
		&msg.Body,
		&msg.CreatedAt,
	)
	if err != nil {
		return Message{}, err
	}
	msg.User = user
	return msg, nil
}

func (r *Repository) GetOrCreateConversation(ctx context.Context, title string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, "select id from conversations where title = $1 and kind = 'group' limit 1", title).Scan(&id)
	if err == nil {
		return id, nil
	}

	err = r.pool.QueryRow(ctx, `
		insert into conversations (kind, title) 
		values ('group', $1) 
		returning id
	`, title).Scan(&id)
	return id, err
}
