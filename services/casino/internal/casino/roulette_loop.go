package casino

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

const rouletteLoopTick = 500 * time.Millisecond

type RouletteLoop struct {
	store *Store
}

func NewRouletteLoop(store *Store) *RouletteLoop {
	return &RouletteLoop{store: store}
}

func (l *RouletteLoop) Start(ctx context.Context) {
	if l == nil || l.store == nil {
		return
	}
	go l.run(ctx)
}

func (l *RouletteLoop) run(ctx context.Context) {
	l.tick(ctx)

	ticker := time.NewTicker(rouletteLoopTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.tick(ctx)
		}
	}
}

func (l *RouletteLoop) tick(ctx context.Context) {
	_, err := l.store.SyncRouletteRound(ctx)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Printf("casino roulette loop: get round: %v", err)
		}
		return
	}
}
