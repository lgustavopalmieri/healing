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

	deliveryChan := make(chan kafkalib.Event)

	err = p.producer.Produce(msg, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce kafka message: %w", err)
	}

	e := <-deliveryChan
	deliveredMsg := e.(*kafkalib.Message)
	if deliveredMsg.TopicPartition.Error != nil {
		return fmt.Errorf("kafka delivery failed: %w", deliveredMsg.TopicPartition.Error)
	}

	return nil
}

func (p *KafkaProducer) Close() {
	p.producer.Flush(10000)
	p.producer.Close()
}
