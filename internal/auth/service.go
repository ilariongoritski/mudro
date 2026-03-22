package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrNoSession          = errors.New("no session found")
)

type User struct {
	ID           int64
	Username     string
	Email        *string
	PasswordHash string
	Role         string
	IsPremium    bool
	TelegramID   *int64
	TelegramName *string
	CreatedAt    time.Time
}

type Service struct {
	pool      *pgxpool.Pool
	jwtSecret []byte
}

func NewService(pool *pgxpool.Pool, secret string) *Service {
	return &Service{
		pool:      pool,
		jwtSecret: []byte(secret),
	}
}

// Register creates a new user.
func (s *Service) Register(ctx context.Context, username, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user User
	err = s.pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, role, created_at)
		VALUES ($1, $2, 'user', NOW())
		RETURNING id, username, role, created_at
	`, username, string(hash)).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "users_username_key") {
			return nil, ErrUserExists
		}
		return nil, err
	}

	return &user, nil
}

// Login validates credentials and returns a user and token.
func (s *Service) Login(ctx context.Context, login, password string) (*User, string, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, role, created_at
		FROM users WHERE username = $1 OR email = $1
	`, login).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	tokenString, err := s.IssueToken(&user)
	if err != nil {
		return nil, "", err
	}

	return &user, tokenString, nil
}

// IssueToken signs and returns a JWT token for an authenticated user.
func (s *Service) IssueToken(user *User) (string, error) {
	if user == nil {
		return "", errors.New("user is nil")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"exp":  time.Now().Add(24 * 7 * time.Hour).Unix(),
		"role": user.Role,
	})

	return token.SignedString(s.jwtSecret)
}

// ValidateToken parses a JWT token and returns standard claims.
func (s *Service) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure token method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// GetUserByID loads user info.
func (s *Service) GetUserByID(ctx context.Context, id int64) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT id, username, role, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoSession
		}
		return nil, err
	}
	s.fillPremiumStatus(ctx, &user)

	return &user, nil
}

// ListUsers returns all users in the system.
func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, email, role, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// CountUsers returns total number of users.
func (s *Service) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	if err := s.pool.QueryRow(ctx, `select count(*) from users`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// CountActiveSubscriptions returns number of currently active subscriptions.
func (s *Service) CountActiveSubscriptions(ctx context.Context) (int64, error) {
	var count int64
	if err := s.pool.QueryRow(ctx, `
		select count(*)
		from user_subscriptions
		where status = 'active' and expires_at > now()
	`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// AddSubscription creates or updates a subscription for a user.
func (s *Service) AddSubscription(ctx context.Context, userID int64, planID string, duration time.Duration) error {
	expiresAt := time.Now().Add(duration)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO user_subscriptions (user_id, plan_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			plan_id = EXCLUDED.plan_id,
			expires_at = EXCLUDED.expires_at,
			status = 'active',
			updated_at = NOW()
	`, userID, planID, expiresAt)
	return err
}

// HasActiveSubscription checks if a user has a valid active subscription.
func (s *Service) HasActiveSubscription(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_subscriptions
			WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
		)
	`, userID).Scan(&exists)
	return exists, err
}

func (s *Service) fillPremiumStatus(ctx context.Context, user *User) {
	if user == nil {
		return
	}
	premium, _ := s.HasActiveSubscription(ctx, user.ID)
	user.IsPremium = premium
}

// FindOrCreateTelegramUser returns an existing Telegram-linked user or creates one.
func (s *Service) FindOrCreateTelegramUser(ctx context.Context, telegramID int64, telegramUsername string) (*User, error) {
	if telegramID <= 0 {
		return nil, errors.New("telegram id is required")
	}

	loaded, err := s.loadUserByTelegramID(ctx, telegramID)
	if err == nil {
		if strings.TrimSpace(telegramUsername) != "" {
			name := strings.TrimSpace(telegramUsername)
			_, _ = s.pool.Exec(ctx, `
				update users
				set telegram_username = $2, updated_at = now()
				where id = $1
			`, loaded.ID, name)
			loaded.TelegramName = &name
		}
		s.fillPremiumStatus(ctx, loaded)
		return loaded, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	baseUsername := normalizeTelegramUsername(telegramUsername, telegramID)
	passwordHash, err := randomPasswordHash()
	if err != nil {
		return nil, err
	}

	for attempt := 0; attempt < 10; attempt++ {
		candidate := baseUsername
		if attempt > 0 {
			candidate = fmt.Sprintf("%s_%d", baseUsername, attempt+1)
		}

		var user User
		telegramName := nullableTrimmed(telegramUsername)
		err = s.pool.QueryRow(ctx, `
			insert into users (username, password_hash, role, created_at, updated_at, telegram_id, telegram_username)
			values ($1, $2, 'user', now(), now(), $3, $4)
			returning id, username, role, created_at, telegram_id, telegram_username
		`, candidate, passwordHash, telegramID, telegramName).Scan(
			&user.ID, &user.Username, &user.Role, &user.CreatedAt, &user.TelegramID, &user.TelegramName,
		)
		if err == nil {
			s.fillPremiumStatus(ctx, &user)
			return &user, nil
		}

		// Username collision can happen when tg username already exists in local users.
		if strings.Contains(err.Error(), "users_username_key") {
			continue
		}

		// Race: another request already created the same telegram user.
		if strings.Contains(err.Error(), "telegram_id") {
			loaded, loadErr := s.loadUserByTelegramID(ctx, telegramID)
			if loadErr != nil {
				return nil, loadErr
			}
			s.fillPremiumStatus(ctx, loaded)
			return loaded, nil
		}

		return nil, err
	}

	return nil, ErrUserExists
}

func (s *Service) loadUserByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		select id, username, email, role, created_at, telegram_id, telegram_username
		from users
		where telegram_id = $1
	`, telegramID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt, &user.TelegramID, &user.TelegramName,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func randomPasswordHash() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(raw)), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func normalizeTelegramUsername(username string, telegramID int64) string {
	trimmed := strings.ToLower(strings.TrimSpace(username))
	if trimmed == "" {
		return fmt.Sprintf("tg_%d", telegramID)
	}

	var b strings.Builder
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	cleaned := b.String()
	if cleaned == "" {
		return fmt.Sprintf("tg_%d", telegramID)
	}
	if len(cleaned) > 32 {
		return cleaned[:32]
	}
	return cleaned
}

func nullableTrimmed(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
