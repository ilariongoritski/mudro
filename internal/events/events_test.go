package events

import (
	"context"
	"testing"
)

func TestNoopPublisher(t *testing.T) {
	p := NoopPublisher{}
	if err := p.PublishTaskEvent(context.Background(), TaskEvent{TaskID: 1, Status: "queued"}); err != nil {
		t.Fatalf("noop publish: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("noop close: %v", err)
	}
}

func TestNewKafkaPublisherValidation(t *testing.T) {
	if _, err := NewKafkaPublisher(nil, "topic", "cid"); err == nil {
		t.Fatal("expected brokers validation error")
	}
	if _, err := NewKafkaPublisher([]string{"localhost:9092"}, "", "cid"); err == nil {
		t.Fatal("expected topic validation error")
	}
}
