package auth

import (
	"context"
	"time"
)

// UserRepository defines the data-access contract for user operations.
// The auth usecase (Service) depends only on this interface — never on pgx directly.
// This satisfies the Dependency Rule: outer layers (repository) depend on inner ones (domain), not vice-versa.
type UserRepository interface {
	// FindByLogin looks up a user by username or email.
	// Returns ErrInvalidCredentials when not found.
	FindByLogin(ctx context.Context, login string) (*User, error)

	// FindByID looks up a user by primary key.
	// Returns ErrNoSession when not found.
	FindByID(ctx context.Context, id int64) (*User, error)

	// FindByTelegramID looks up a user by Telegram ID.
	// Returns the raw pgx error (including pgx.ErrNoRows) to allow business-logic branching in Service.
	FindByTelegramID(ctx context.Context, telegramID int64) (*User, error)

	// Create inserts a new standard user.
	// Returns ErrUserExists on unique-constraint violation.
	Create(ctx context.Context, username, email, passwordHash string) (*User, error)

	// CreateFromTelegram inserts a Telegram-linked user.
	// Returns ErrUserExists on username collision, ErrTelegramIDConflict on telegram_id collision.
	CreateFromTelegram(ctx context.Context, params TelegramUserParams) (*User, error)

	// UpdateTelegramName updates the display name for a Telegram user.
	UpdateTelegramName(ctx context.Context, id int64, name string) error

	// HasActiveSubscription reports whether a user has a non-expired active subscription.
	HasActiveSubscription(ctx context.Context, userID int64) (bool, error)

	// ListAll returns all users ordered by creation date descending.
	ListAll(ctx context.Context) ([]User, error)

	// Count returns the total number of users.
	Count(ctx context.Context) (int64, error)

	// CountActiveSubscriptions returns the number of currently active subscriptions.
	CountActiveSubscriptions(ctx context.Context) (int64, error)

	// AddSubscription creates or updates a subscription for a user.
	AddSubscription(ctx context.Context, userID int64, planID string, duration time.Duration) error
}

// TelegramUserParams holds the fields needed to create a Telegram-linked user.
type TelegramUserParams struct {
	Username     string
	PasswordHash string
	TelegramID   int64
	TelegramName *string
}

// ErrTelegramIDConflict is returned when a race condition causes a duplicate telegram_id insert.
var ErrTelegramIDConflict = errTelegramIDConflict{}

type errTelegramIDConflict struct{}

func (errTelegramIDConflict) Error() string { return "telegram id conflict" }
