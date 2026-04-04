package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgUserRepository is the PostgreSQL implementation of UserRepository.
type pgUserRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository returns a production Postgres-backed UserRepository.
func NewPgRepository(pool *pgxpool.Pool) UserRepository {
	return &pgUserRepository{pool: pool}
}

func (r *pgUserRepository) FindByLogin(ctx context.Context, login string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, role, coalesce(avatar_url, ''), created_at
		FROM users WHERE username = $1 OR email = $1
	`, login).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) FindByID(ctx context.Context, id int64) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, email, role, coalesce(avatar_url, ''), created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoSession
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) FindByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, email, role, coalesce(avatar_url, ''), created_at, telegram_id, telegram_username
		FROM users WHERE telegram_id = $1
	`, telegramID).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.AvatarURL, &u.CreatedAt, &u.TelegramID, &u.TelegramName)
	if err != nil {
		return nil, err // pass pgx.ErrNoRows through — Service uses it for branch logic
	}
	return &u, nil
}

func (r *pgUserRepository) Create(ctx context.Context, username, email, passwordHash string) (*User, error) {
	var u User
	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, 'user', NOW(), NOW())
		RETURNING id, username, email, role, coalesce(avatar_url, ''), created_at
	`, username, emailPtr, passwordHash).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "users_username_key") || strings.Contains(err.Error(), "users_email_key") {
			return nil, ErrUserExists
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) CreateFromTelegram(ctx context.Context, params TelegramUserParams) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, role, created_at, updated_at, telegram_id, telegram_username)
		VALUES ($1, $2, 'user', NOW(), NOW(), $3, $4)
		RETURNING id, username, role, coalesce(avatar_url, ''), created_at, telegram_id, telegram_username
	`, params.Username, params.PasswordHash, params.TelegramID, params.TelegramName).Scan(
		&u.ID, &u.Username, &u.Role, &u.AvatarURL, &u.CreatedAt, &u.TelegramID, &u.TelegramName,
	)
	if err != nil {
		if strings.Contains(err.Error(), "users_username_key") {
			return nil, ErrUserExists
		}
		if strings.Contains(err.Error(), "telegram_id") {
			return nil, ErrTelegramIDConflict
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) UpdateTelegramName(ctx context.Context, id int64, name string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET telegram_username = $2, updated_at = NOW() WHERE id = $1
	`, id, name)
	return err
}

func (r *pgUserRepository) HasActiveSubscription(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_subscriptions
			WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
		)
	`, userID).Scan(&exists)
	return exists, err
}

func (r *pgUserRepository) ListAll(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, username, email, role, coalesce(avatar_url, ''), created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.AvatarURL, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *pgUserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *pgUserRepository) CountActiveSubscriptions(ctx context.Context) (int64, error) {
	var count int64
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_subscriptions
		WHERE status = 'active' AND expires_at > NOW()
	`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *pgUserRepository) AddSubscription(ctx context.Context, userID int64, planID string, duration time.Duration) error {
	expiresAt := time.Now().Add(duration)
	_, err := r.pool.Exec(ctx, `
		WITH updated AS (
			UPDATE user_subscriptions
			SET plan_id = $2,
				expires_at = $3,
				status = 'active',
				updated_at = NOW()
			WHERE user_id = $1
			RETURNING 1
		)
		INSERT INTO user_subscriptions (user_id, plan_id, status, expires_at)
		SELECT $1, $2, 'active', $3
		WHERE NOT EXISTS (SELECT 1 FROM updated)
	`, userID, planID, expiresAt)
	return err
}
