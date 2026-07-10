package profile

import "time"

// UserProfile represents the profile fields on top of users table.
type UserProfile struct {
	ID                  int64     `json:"id" db:"id"`
	DisplayName         string    `json:"display_name" db:"display_name"`
	Username            string    `json:"username" db:"username"`
	Email               *string   `json:"email,omitempty" db:"email"`
	Age                 *int      `json:"age,omitempty" db:"age"`
	Bio                 *string   `json:"bio,omitempty" db:"bio"`
	SocialLinks         JSONMap   `json:"social_links,omitempty" db:"social_links"`
	AvatarURL           *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	ProfileCompletion   int       `json:"profile_completion" db:"profile_completion"`
	Rating              int       `json:"rating" db:"rating"`
	TelegramID          *int64    `json:"telegram_id,omitempty" db:"telegram_id"`
	TelegramUsername    *string   `json:"telegram_username,omitempty" db:"telegram_username"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// JSONMap for social_links
type JSONMap map[string]string

// Activity log entry
type Activity struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"`
	RefID     *int64    `json:"ref_id,omitempty" db:"ref_id"`
	Metadata  JSONMap   `json:"metadata,omitempty" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
