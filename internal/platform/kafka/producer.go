package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	kafkalib "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

type KafkaProducer struct {
	producer *kafkalib.Producer
}

func NewKafkaProducer(config *kafkalib.ConfigMap) (*KafkaProducer, error) {
	producer, err := kafkalib.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &KafkaProducer{producer: producer}, nil
}

func (p *KafkaProducer) Dispatch(ctx context.Context, evt event.Event) error {
	value, err := json.Marshal(evt.Payload)
	if err != nil {
		return fmt.Errorf("error serializing event payload: %w", err)
	}

	topic := evt.Name

	msg := &kafkalib.Message{
		TopicPartition: kafkalib.TopicPartition{
			Topic:     &topic,
			Partition: kafkalib.PartitionAny,
		},
		Value:     value,
		Timestamp: evt.Timestamp,
	}

	err = p.producer.Produce(msg, nil)
	if err != nil {
		return fmt.Errorf("failed to produce kafka message: %w", err)
	}
	fmt.Println("Kafka message produced: ", msg)

	return nil
}

func (p *KafkaProducer) Close() {
	p.producer.Flush(10000)
	p.producer.Close()
}
