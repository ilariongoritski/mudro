package casino

import (
	"context"
	"sync"
	"time"
)

const hubBroadcastInterval = time.Second

// RouletteHub maintains a single background goroutine that polls roulette state
// once per second and fans it out to all connected SSE subscribers.
// This replaces the previous O(N) pattern where each SSE client triggered its
// own DB query every second.
type RouletteHub struct {
	store *Store

	mu   sync.Mutex
	subs map[chan RouletteState]struct{}
}

func NewRouletteHub(store *Store) *RouletteHub {
	return &RouletteHub{
		store: store,
		subs:  make(map[chan RouletteState]struct{}),
	}
}

// Start launches the broadcast loop; call once, pass the service-lifetime context.
func (h *RouletteHub) Start(ctx context.Context) {
	go h.run(ctx)
}

// Subscribe returns a buffered channel that receives state broadcasts.
// Call Unsubscribe when the SSE connection closes.
func (h *RouletteHub) Subscribe() chan RouletteState {
	ch := make(chan RouletteState, 2)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes the channel from the broadcast set and closes it.
func (h *RouletteHub) Unsubscribe(ch chan RouletteState) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

func (h *RouletteHub) run(ctx context.Context) {
	h.broadcast(ctx)

	ticker := time.NewTicker(hubBroadcastInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.broadcast(ctx)
		}
	}
}

// broadcast fetches the shared state (no user-specific bets) and delivers it
// to all subscribers. Slow subscribers are skipped; they receive the next tick.
func (h *RouletteHub) broadcast(ctx context.Context) {
	// userID=0 → no per-user bets in the shared payload
	state, err := h.store.GetCurrentRouletteState(ctx, 0)
	if err != nil {
		return
	}
	state.MyBets = nil // guard: never leak user data in shared broadcast

	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- state:
		default:
			// subscriber's buffer full — skip; it will catch the next tick
		}
	}
}
