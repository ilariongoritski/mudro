package profile

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Service handles profile logic.
type Service struct {
	db *sql.DB
}

// NewService creates profile service.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// CalculateCompletion computes 0-100 based on filled fields (name+username = 20 total).
func (s *Service) CalculateCompletion(p *UserProfile) int {
	score := 0
	if p.DisplayName != "" {
		score += 10
	}
	if p.Username != "" {
		score += 10
	}
	if p.Email != nil && *p.Email != "" {
		score += 15
	}
	if p.Age != nil && *p.Age >= 13 {
		score += 15
	}
	if p.Bio != nil && *p.Bio != "" {
		score += 20
	}
	if len(p.SocialLinks) > 0 {
		score += 15
	}
	if p.AvatarURL != nil && *p.AvatarURL != "" {
		score += 15
	}
	if score > 100 {
		score = 100
	}
	return score
}

// UpdateRating recalculates rating (base 20 from name/username + activity bonus ~30%).
// In real use, call after profile update and activity logging.
func (s *Service) UpdateRating(ctx context.Context, userID int64, base int, activityBonus int) error {
	rating := base + activityBonus
	if rating < 0 {
		rating = 0
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET rating = $1, updated_at = now() WHERE id = $2
	`, rating, userID)
	return err
}

// LogActivity records any user action.
func (s *Service) LogActivity(ctx context.Context, userID int64, typ string, refID *int64, meta JSONMap) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_activities (user_id, type, ref_id, metadata) 
		VALUES ($1, $2, $3, $4)
	`, userID, typ, refID, meta)
	return err
}

// GetProfile fetches profile + basic casino stats (balance, games, max_win).
func (s *Service) GetProfile(ctx context.Context, userID int64) (*UserProfile, error) {
	var p UserProfile
	err := s.db.QueryRowContext(ctx, `
		SELECT id, display_name, username, email, age, bio, social_links, avatar_url,
		       profile_completion, rating, telegram_id, telegram_username, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&p.ID, &p.DisplayName, &p.Username, &p.Email, &p.Age, &p.Bio, &p.SocialLinks, &p.AvatarURL,
		&p.ProfileCompletion, &p.Rating, &p.TelegramID, &p.TelegramUsername, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
