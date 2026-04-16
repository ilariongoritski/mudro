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

type Message struct {
	ID             int64        `json:"id"`
	ConversationID int64        `json:"conversation_id"`
	SenderID       int64        `json:"sender_id"`
	Body           string       `json:"body"`
	EncryptedBody  *string      `json:"encrypted_body,omitempty"`
	Nonce          *string      `json:"nonce,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	User           UserIdentity `json:"user"`
	Room           string       `json:"room,omitempty"`
}

type UserIdentity struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	AvatarURL string `json:"avatar_url"`
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
			m.encrypted_body,
			m.nonce,
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
			&msg.EncryptedBody,
			&msg.Nonce,
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

func (r *Repository) InsertMessage(ctx context.Context, user UserIdentity, room, body string, encryptedBody, nonce *string) (Message, error) {
	room = normalizeRoom(room)
	convID, err := r.GetOrCreateConversation(ctx, room)
	if err != nil {
		return Message{}, err
	}

	var msg Message
	msg.Room = room
	err = r.pool.QueryRow(ctx, `
		insert into messages (conversation_id, sender_id, body, encrypted_body, nonce)
		values ($1, $2, $3, $4, $5)
		returning id, conversation_id, sender_id, body, encrypted_body, nonce, created_at
	`, convID, user.ID, body, encryptedBody, nonce).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderID,
		&msg.Body,
		&msg.EncryptedBody,
		&msg.Nonce,
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

func (r *Repository) SaveUserKeys(ctx context.Context, userID int64, identityKey, signedPrekey, signature string, oneTimePrekeys []map[string]any) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		insert into user_keys (user_id, identity_key, signed_prekey, prekey_signature, updated_at)
		values ($1, $2, $3, $4, now())
		on conflict (user_id) do update set
			identity_key = excluded.identity_key,
			signed_prekey = excluded.signed_prekey,
			prekey_signature = excluded.prekey_signature,
			updated_at = now()
	`, userID, identityKey, signedPrekey, signature)
	if err != nil {
		return err
	}

	for _, pk := range oneTimePrekeys {
		keyID := pk["id"].(float64)
		pubKey := pk["key"].(string)
		_, err = tx.Exec(ctx, `
			insert into one_time_prekeys (user_id, key_id, public_key)
			values ($1, $2, $3)
			on conflict (user_id, key_id) do update set public_key = excluded.public_key
		`, userID, int(keyID), pubKey)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

type UserKeysBundle struct {
	IdentityKey    string `json:"identity_key"`
	SignedPrekey   string `json:"signed_prekey"`
	Signature      string `json:"signature"`
	OneTimePrekey  *OneTimePrekey `json:"one_time_prekey,omitempty"`
}

type OneTimePrekey struct {
	ID  int    `json:"id"`
	Key string `json:"key"`
}

func (r *Repository) GetUserKeysBundle(ctx context.Context, userID int64) (*UserKeysBundle, error) {
	bundle := &UserKeysBundle{}
	err := r.pool.QueryRow(ctx, `
		select identity_key, signed_prekey, prekey_signature
		from user_keys where user_id = $1
	`, userID).Scan(&bundle.IdentityKey, &bundle.SignedPrekey, &bundle.Signature)
	if err != nil {
		return nil, err
	}

	var otpk OneTimePrekey
	err = r.pool.QueryRow(ctx, `
		delete from one_time_prekeys
		where id = (
			select id from one_time_prekeys
			where user_id = $1
			limit 1
			for update skip locked
		)
		returning key_id, public_key
	`, userID).Scan(&otpk.ID, &otpk.Key)
	if err == nil {
		bundle.OneTimePrekey = &otpk
	}

	return bundle, nil
}
