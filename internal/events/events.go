package events

import (
	"context"
	"time"
)

type Publisher interface {
	PublishTaskEvent(ctx context.Context, ev TaskEvent) error
	Close() error
}

type TaskEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	TaskID    int64     `json:"task_id"`
	Kind      string    `json:"kind,omitempty"`
	Status    string    `json:"status"`
	DedupeKey string    `json:"dedupe_key,omitempty"`
	Error     string    `json:"error,omitempty"`
	Occurred  time.Time `json:"occurred_at"`
}

type NoopPublisher struct{}

func (NoopPublisher) PublishTaskEvent(context.Context, TaskEvent) error { return nil }
func (NoopPublisher) Close() error                                      { return nil }
