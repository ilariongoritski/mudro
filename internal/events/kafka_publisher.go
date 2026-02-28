package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	w *kafka.Writer
}

func NewKafkaPublisher(brokers []string, topic, clientID string) (*KafkaPublisher, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are empty")
	}
	if topic == "" {
		return nil, fmt.Errorf("kafka topic is empty")
	}
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		AllowAutoTopicCreation: true,
		RequiredAcks:           kafka.RequireOne,
		BatchTimeout:           50 * time.Millisecond,
		Async:                  false,
		Transport: &kafka.Transport{
			ClientID: clientID,
		},
	}
	return &KafkaPublisher{w: w}, nil
}

func (p *KafkaPublisher) PublishTaskEvent(ctx context.Context, ev TaskEvent) error {
	if ev.EventID == "" {
		ev.EventID = fmt.Sprintf("%d-%d", ev.TaskID, time.Now().UnixNano())
	}
	if ev.Occurred.IsZero() {
		ev.Occurred = time.Now().UTC()
	}
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	key := strconv.FormatInt(ev.TaskID, 10)
	return p.w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: b,
		Time:  ev.Occurred,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(ev.EventType)},
			{Key: "status", Value: []byte(ev.Status)},
		},
	})
}

func (p *KafkaPublisher) Close() error {
	if p == nil || p.w == nil {
		return nil
	}
	return p.w.Close()
}
