package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaProducer struct {
	client *kgo.Client
}

func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &KafkaProducer{client: client}, nil
}

func (p *KafkaProducer) Dispatch(ctx context.Context, evt event.Event) error {
	value, err := json.Marshal(evt.Payload)
	if err != nil {
		return fmt.Errorf("error serializing event payload: %w", err)
	}

	record := &kgo.Record{
		Topic:     evt.Name,
		Value:     value,
		Timestamp: evt.Timestamp,
	}

	result := p.client.ProduceSync(ctx, record)
	if err := result.FirstErr(); err != nil {
		return fmt.Errorf("kafka delivery failed: %w", err)
	}

	return nil
}

func (p *KafkaProducer) Close() {
	p.client.Close()
}
