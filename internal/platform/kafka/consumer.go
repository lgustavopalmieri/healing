package kafka

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

// repo pocs/kafka-poc

type KafkaConsumer struct {
	consumer *kafka.Consumer
	manager  *event.ListenerManager
}

func NewKafkaConsumer(config *kafka.ConfigMap, manager *event.ListenerManager) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, err
	}

	err = c.SubscribeTopics(manager.Topics(), nil)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer: c,
		manager:  manager,
	}, nil
}

func (dc *KafkaConsumer) Start(ctx context.Context) {
	for {
		msg, err := dc.consumer.ReadMessage(-1)
		if err != nil {
			fmt.Printf("Kafka error: %v\n", err)
			continue
		}

		topic := *msg.TopicPartition.Topic
		listener, ok := dc.manager.GetListener(topic)
		if !ok {
			fmt.Printf("No listener registered for topic: %s\n", topic)
			continue
		}

		evt := event.NewEvent(*msg.TopicPartition.Topic, msg.Value)

		if err := listener.Handle(ctx, evt); err != nil {
			fmt.Printf("Error handling event on topic %s: %v\n", topic, err)
			continue
		}

		_, err = dc.consumer.CommitMessage(msg)
		if err != nil {
			fmt.Printf("Failed to commit message: %v\n", err)
		} else {
			fmt.Printf("✅ [%s] Offset comitado para %s [%d]@%d\n",
				dc.consumer.GetRebalanceProtocol(),
				topic,
				msg.TopicPartition.Partition,
				msg.TopicPartition.Offset)
		}
	}
}