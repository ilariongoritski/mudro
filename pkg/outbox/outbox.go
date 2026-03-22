package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// Event represents a single domain event in canonical envelope form.
type Event struct {
	EventID      string          `json:"event_id"`
	EventType    string          `json:"event_type"`
	EventVersion int             `json:"event_version"`
	Producer     string          `json:"producer"`
	AggregateID  string          `json:"aggregate_id"`
	OccurredAt   time.Time       `json:"occurred_at"`
	TraceID      string          `json:"trace_id,omitempty"`
	DedupeKey    string          `json:"dedupe_key,omitempty"`
	Payload      json.RawMessage `json:"payload"`
}

// Record is a DB outbox row waiting to be published.
type Record struct {
	ID          int64
	Event       Event
	PublishedAt *time.Time
}

var (
	ErrStoreNotConfigured     = errors.New("outbox store is not configured")
	ErrPublisherNotConfigured = errors.New("outbox publisher is not configured")
)

// Store abstracts persistence for transactional outbox rows.
type Store interface {
	Insert(ctx context.Context, event Event) (int64, error)
	LoadPending(ctx context.Context, limit int) ([]Record, error)
	MarkPublished(ctx context.Context, id int64, publishedAt time.Time) error
}

// Publisher abstracts event transport (for example Kafka/NATS).
type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

// Service contains orchestration logic for outbox delivery.
// This is a Wave-0 skeleton and intentionally transport/storage-agnostic.
type Service struct {
	store     Store
	publisher Publisher
	now       func() time.Time
}

func NewService(store Store, publisher Publisher) *Service {
	return &Service{
		store:     store,
		publisher: publisher,
		now:       time.Now().UTC,
	}
}

// Enqueue persists an event in the outbox store. Caller should invoke this
// inside the same transaction as domain writes in a concrete implementation.
func (s *Service) Enqueue(ctx context.Context, event Event) (int64, error) {
	if s == nil || s.store == nil {
		return 0, ErrStoreNotConfigured
	}
	return s.store.Insert(ctx, event)
}

// FlushOnce publishes up to `limit` pending events and marks each row published.
func (s *Service) FlushOnce(ctx context.Context, limit int) (int, error) {
	if s == nil || s.store == nil {
		return 0, ErrStoreNotConfigured
	}
	if s.publisher == nil {
		return 0, ErrPublisherNotConfigured
	}
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.store.LoadPending(ctx, limit)
	if err != nil {
		return 0, err
	}

	processed := 0
	for _, row := range rows {
		if err := s.publisher.Publish(ctx, row.Event); err != nil {
			return processed, err
		}
		if err := s.store.MarkPublished(ctx, row.ID, s.now()); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}
