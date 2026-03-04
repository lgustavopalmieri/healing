package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	platformkafka "github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
)

type KafkaContainer struct {
	Container testcontainers.Container
	Brokers   []string
}

func SetupKafkaContainer(t *testing.T) *KafkaContainer {
	ctx := context.Background()

	kafkaContainer, err := tckafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		tckafka.WithClusterID("test-cluster"),
	)
	require.NoError(t, err)

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)

	return &KafkaContainer{
		Container: kafkaContainer,
		Brokers:   brokers,
	}
}

func (c *KafkaContainer) Terminate(t *testing.T) {
	ctx := context.Background()
	err := c.Container.Terminate(ctx)
	require.NoError(t, err)
}

func (c *KafkaContainer) BootstrapServers() string {
	return c.Brokers[0]
}

func (c *KafkaContainer) CreateProducer(t *testing.T) *platformkafka.KafkaProducer {
	producer, err := platformkafka.NewKafkaProducer(c.Brokers)
	require.NoError(t, err)
	return producer
}

func (c *KafkaContainer) CreateConsumer(t *testing.T, groupID string, manager *event.ListenerManager) *platformkafka.KafkaConsumer {
	consumer, err := platformkafka.NewKafkaConsumer(c.Brokers, groupID, manager)
	require.NoError(t, err)
	return consumer
}

func (c *KafkaContainer) ProduceEvent(t *testing.T, topic string, payload any) {
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	client, err := kgo.NewClient(
		kgo.SeedBrokers(c.Brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	require.NoError(t, err)
	defer client.Close()

	record := &kgo.Record{
		Topic: topic,
		Value: data,
	}

	result := client.ProduceSync(context.Background(), record)
	require.NoError(t, result.FirstErr())
}

func (c *KafkaContainer) ConsumeEvent(t *testing.T, topic string, groupID string, timeout time.Duration) []byte {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(c.Brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		fetches := client.PollFetches(ctx)
		if ctx.Err() != nil {
			t.Fatalf("timeout waiting for message on topic %s", topic)
			return nil
		}

		var result []byte
		fetches.EachRecord(func(record *kgo.Record) {
			if result == nil {
				result = record.Value
			}
		})

		if result != nil {
			return result
		}
	}
}

type TestHelper struct {
	SharedContainer *KafkaContainer
}

func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

func (h *TestHelper) RunTestMain(m *testing.M) {
	h.SharedContainer = SetupKafkaContainer(&testing.T{})

	code := m.Run()

	if h.SharedContainer != nil {
		h.SharedContainer.Terminate(&testing.T{})
	}

	os.Exit(code)
}

func (c *KafkaContainer) CreateTopics(topics ...string) error {
	client, err := kgo.NewClient(kgo.SeedBrokers(c.Brokers...))
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}
	defer client.Close()

	admin := kadm.NewClient(client)
	defer admin.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := admin.CreateTopics(ctx, 1, 1, nil, topics...)
	if err != nil {
		return fmt.Errorf("failed to create topics: %w", err)
	}

	for _, r := range resp.Sorted() {
		if r.Err != nil {
			if r.Err.Error() != "TOPIC_ALREADY_EXISTS" {
				return fmt.Errorf("failed to create topic %s: %v", r.Topic, r.Err)
			}
		}
	}

	time.Sleep(2 * time.Second)
	return nil
}

func WaitForConsumerGroupReady(bootstrapServers string, groupID string, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	client, err := kgo.NewClient(kgo.SeedBrokers(bootstrapServers))
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}
	defer client.Close()

	admin := kadm.NewClient(client)
	defer admin.Close()

	for {
		select {
		case <-deadline:
			return fmt.Errorf("timeout waiting for consumer group %s to be ready", groupID)
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			groups, err := admin.ListGroups(ctx)
			cancel()
			if err != nil {
				continue
			}
			for _, g := range groups.Sorted() {
				if g.Group == groupID {
					return nil
				}
			}
		}
	}
}
