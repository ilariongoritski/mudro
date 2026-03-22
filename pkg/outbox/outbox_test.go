package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type memStore struct {
	rows []Record
}

func (m *memStore) Insert(_ context.Context, event Event) (int64, error) {
	id := int64(len(m.rows) + 1)
	m.rows = append(m.rows, Record{ID: id, Event: event})
	return id, nil
}

func (m *memStore) LoadPending(_ context.Context, limit int) ([]Record, error) {
	out := make([]Record, 0, limit)
	for _, row := range m.rows {
		if row.PublishedAt == nil {
			out = append(out, row)
		}
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *memStore) MarkPublished(_ context.Context, id int64, at time.Time) error {
	for i := range m.rows {
		if m.rows[i].ID == id {
			m.rows[i].PublishedAt = &at
			return nil
		}
	}
	return errors.New("row not found")
}

type memPublisher struct {
	events []Event
}

func (m *memPublisher) Publish(_ context.Context, event Event) error {
	m.events = append(m.events, event)
	return nil
}

func TestFlushOncePublishesAndMarksRows(t *testing.T) {
	store := &memStore{}
	pub := &memPublisher{}
	svc := NewService(store, pub)

	payload := json.RawMessage(`{"status":"ok"}`)
	_, err := svc.Enqueue(context.Background(), Event{
		EventID:      "evt-1",
		EventType:    "mudro.agent.task.v1.created",
		EventVersion: 1,
		Producer:     "agent-service",
		AggregateID:  "task-1",
		OccurredAt:   time.Now().UTC(),
		Payload:      payload,
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	n, err := svc.FlushOnce(context.Background(), 10)
	if err != nil {
		t.Fatalf("flush: %v", err)
	}
	if n != 1 {
		t.Fatalf("processed = %d, want 1", n)
	}
	if len(pub.events) != 1 {
		t.Fatalf("published = %d, want 1", len(pub.events))
	}
	if store.rows[0].PublishedAt == nil {
		t.Fatal("row is not marked as published")
	}
}
