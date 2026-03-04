package kafka

import (
	"context"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaConsumer struct {
	client  *kgo.Client
	manager *event.ListenerManager
}

func NewKafkaConsumer(brokers []string, groupID string, manager *event.ListenerManager, opts ...kgo.Opt) (*KafkaConsumer, error) {
	baseOpts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(manager.Topics()...),
		kgo.DisableAutoCommit(),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	}
	baseOpts = append(baseOpts, opts...)

	client, err := kgo.NewClient(baseOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	return &KafkaConsumer{
		client:  client,
		manager: manager,
	}, nil
}

func (dc *KafkaConsumer) Start(ctx context.Context) {
	defer dc.client.Close()

	for {
		fetches := dc.client.PollFetches(ctx)
		if ctx.Err() != nil {
			fmt.Println("Kafka consumer stopped: context cancelled")
			return
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				fmt.Printf("Kafka error: topic=%s partition=%d err=%v\n", e.Topic, e.Partition, e.Err)
			}
		}

		fetches.EachRecord(func(record *kgo.Record) {
			topic := record.Topic
			listener, ok := dc.manager.GetListener(topic)
			if !ok {
				fmt.Printf("No listener registered for topic: %s\n", topic)
				return
			}

			evt := event.NewEvent(topic, record.Value)

			if err := listener.Handle(ctx, evt); err != nil {
				fmt.Printf("Error handling event on topic %s: %v\n", topic, err)
				return
			}

			if err := dc.client.CommitRecords(ctx, record); err != nil {
				fmt.Printf("Failed to commit message: %v\n", err)
			} else {
				fmt.Printf("✅ Offset committed for %s [%d]@%d\n",
					topic,
					record.Partition,
					record.Offset)
			}
		})
	}
}
