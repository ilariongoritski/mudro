package casino

import (
	"context"
	"fmt"
	"strings"
)

func (s *Store) RotateServerSeed(ctx context.Context, userID int64) (string, error) {
	newSeed, err := GenerateServerSeed()
	if err != nil {
		return "", err
	}
	newHash := HashServerSeed(newSeed)

	_, err = s.pool.Exec(ctx, `
		update casino_players
		set server_seed = $2,
			server_seed_hash = $3,
			current_nonce = 0,
			updated_at = now()
		where user_id = $1
	`, userID, newSeed, newHash)
	return newHash, err
}

func (s *Store) UpdateClientSeed(ctx context.Context, userID int64, newSeed string) error {
	newSeed = strings.TrimSpace(newSeed)
	if newSeed == "" {
		return fmt.Errorf("client seed cannot be empty")
	}
	if len(newSeed) > 64 {
		return fmt.Errorf("client seed too long")
	}

	_, err := s.pool.Exec(ctx, `
		update casino_players
		set client_seed = $2,
			current_nonce = 0,
			updated_at = now()
		where user_id = $1
	`, userID, newSeed)
	return err
}
