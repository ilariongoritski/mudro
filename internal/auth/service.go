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
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrNoSession          = errors.New("no session found")
)

// User is the domain model for an authenticated user.
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

// Service is the auth usecase layer.
// It depends only on UserRepository — never on pgx or pgxpool directly.
type Service struct {
	repo      UserRepository
	jwtSecret []byte
}

// NewService constructs an auth Service with the given repository and JWT secret.
func NewService(repo UserRepository, secret string) *Service {
	return &Service{repo: repo, jwtSecret: []byte(secret)}
}

// Register creates a new user with the given credentials.
func (s *Service) Register(ctx context.Context, username, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, username, email, string(hash))
}

// Login validates credentials and returns the user and a signed JWT token.
func (s *Service) Login(ctx context.Context, login, password string) (*User, string, error) {
	user, err := s.repo.FindByLogin(ctx, login)
	if err != nil {
		return nil, "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}
	tokenString, err := s.IssueToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, tokenString, nil
}

// IssueToken signs and returns a JWT for the given user.
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

// ValidateToken parses a JWT and returns its claims.
func (s *Service) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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

// GetUserByID loads a user and fills their premium status.
func (s *Service) GetUserByID(ctx context.Context, id int64) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.fillPremiumStatus(ctx, user)
	return user, nil
}

// ListUsers returns all users.
func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	return s.repo.ListAll(ctx)
}

// CountUsers returns total number of users.
func (s *Service) CountUsers(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

// CountActiveSubscriptions returns number of currently active subscriptions.
func (s *Service) CountActiveSubscriptions(ctx context.Context) (int64, error) {
	return s.repo.CountActiveSubscriptions(ctx)
}

// AddSubscription creates or updates a subscription for a user.
func (s *Service) AddSubscription(ctx context.Context, userID int64, planID string, duration time.Duration) error {
	return s.repo.AddSubscription(ctx, userID, planID, duration)
}

// HasActiveSubscription checks if a user has a valid active subscription.
func (s *Service) HasActiveSubscription(ctx context.Context, userID int64) (bool, error) {
	return s.repo.HasActiveSubscription(ctx, userID)
}

// FindOrCreateTelegramUser returns an existing Telegram-linked user or creates one.
// The retry loop handles username collisions and telegram_id race conditions.
func (s *Service) FindOrCreateTelegramUser(ctx context.Context, telegramID int64, telegramUsername string) (*User, error) {
	if telegramID <= 0 {
		return nil, errors.New("telegram id is required")
	}

	loaded, err := s.repo.FindByTelegramID(ctx, telegramID)
	if err == nil {
		if name := strings.TrimSpace(telegramUsername); name != "" {
			_ = s.repo.UpdateTelegramName(ctx, loaded.ID, name)
			loaded.TelegramName = &name
		}
		s.fillPremiumStatus(ctx, loaded)
		return loaded, nil
	}
	// If not ErrNoRows — real DB error
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

		telegramName := nullableTrimmed(telegramUsername)
		user, err := s.repo.CreateFromTelegram(ctx, TelegramUserParams{
			Username:     candidate,
			PasswordHash: passwordHash,
			TelegramID:   telegramID,
			TelegramName: telegramName,
		})
		if err == nil {
			s.fillPremiumStatus(ctx, user)
			return user, nil
		}
		if errors.Is(err, ErrUserExists) {
			continue // username collision — try next candidate
		}
		if errors.Is(err, ErrTelegramIDConflict) {
			// Race: another request created the same telegram user
			loaded, loadErr := s.repo.FindByTelegramID(ctx, telegramID)
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

func (s *Service) fillPremiumStatus(ctx context.Context, user *User) {
	if user == nil {
		return
	}
	premium, _ := s.repo.HasActiveSubscription(ctx, user.ID)
	user.IsPremium = premium
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
