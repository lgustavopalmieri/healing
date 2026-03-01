package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	kafkalib "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"

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
	config := &kafkalib.ConfigMap{
		"bootstrap.servers": c.BootstrapServers(),
	}

	producer, err := platformkafka.NewKafkaProducer(config)
	require.NoError(t, err)

	return producer
}

func (c *KafkaContainer) CreateConsumer(t *testing.T, groupID string, manager *event.ListenerManager) *platformkafka.KafkaConsumer {
	config := &kafkalib.ConfigMap{
		"bootstrap.servers":  c.BootstrapServers(),
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	consumer, err := platformkafka.NewKafkaConsumer(config, manager)
	require.NoError(t, err)

	return consumer
}

func (c *KafkaContainer) ProduceEvent(t *testing.T, topic string, payload any) {
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	config := &kafkalib.ConfigMap{
		"bootstrap.servers": c.BootstrapServers(),
	}

	producer, err := kafkalib.NewProducer(config)
	require.NoError(t, err)
	defer producer.Close()

	deliveryChan := make(chan kafkalib.Event)

	err = producer.Produce(&kafkalib.Message{
		TopicPartition: kafkalib.TopicPartition{
			Topic:     &topic,
			Partition: kafkalib.PartitionAny,
		},
		Value: data,
	}, deliveryChan)
	require.NoError(t, err)

	select {
	case e := <-deliveryChan:
		msg := e.(*kafkalib.Message)
		require.NoError(t, msg.TopicPartition.Error)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for kafka message delivery")
	}
}

func (c *KafkaContainer) ConsumeEvent(t *testing.T, topic string, groupID string, timeout time.Duration) []byte {
	config := &kafkalib.ConfigMap{
		"bootstrap.servers":  c.BootstrapServers(),
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
	}

	consumer, err := kafkalib.NewConsumer(config)
	require.NoError(t, err)
	defer consumer.Close()

	err = consumer.SubscribeTopics([]string{topic}, nil)
	require.NoError(t, err)

	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for message on topic %s", topic)
			return nil
		default:
			msg, err := consumer.ReadMessage(500 * time.Millisecond)
			if err != nil {
				continue
			}
			return msg.Value
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
	adminConfig := &kafkalib.ConfigMap{
		"bootstrap.servers": c.BootstrapServers(),
	}

	admin, err := kafkalib.NewAdminClient(adminConfig)
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}
	defer admin.Close()

	topicSpecs := make([]kafkalib.TopicSpecification, len(topics))
	for i, topic := range topics {
		topicSpecs[i] = kafkalib.TopicSpecification{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := admin.CreateTopics(ctx, topicSpecs)
	if err != nil {
		return fmt.Errorf("failed to create topics: %w", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafkalib.ErrNoError && result.Error.Code() != kafkalib.ErrTopicAlreadyExists {
			return fmt.Errorf("failed to create topic %s: %v", result.Topic, result.Error)
		}
	}

	time.Sleep(2 * time.Second)
	return nil
}

func WaitForConsumerGroupReady(bootstrapServers string, groupID string, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	adminConfig := &kafkalib.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	}

	admin, err := kafkalib.NewAdminClient(adminConfig)
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}
	defer admin.Close()

	for {
		select {
		case <-deadline:
			return fmt.Errorf("timeout waiting for consumer group %s to be ready", groupID)
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			result, err := admin.ListConsumerGroups(ctx, kafkalib.SetAdminRequestTimeout(5*time.Second))
			cancel()
			if err != nil {
				continue
			}
			for _, group := range result.Valid {
				if group.GroupID == groupID {
					return nil
				}
			}
		}
	}
}
