package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func HealthCheck(ctx context.Context, brokers []string, opts ...kgo.Opt) error {
	baseOpts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
	}
	baseOpts = append(baseOpts, opts...)

	client, err := kgo.NewClient(baseOpts...)
	if err != nil {
		return fmt.Errorf("failed to create kafka health check client: %w", err)
	}
	defer client.Close()

	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(checkCtx); err != nil {
		return fmt.Errorf("kafka ping failed: %w", err)
	}

	admin := kadm.NewClient(client)
	metadata, err := admin.BrokerMetadata(checkCtx)
	if err != nil {
		return fmt.Errorf("failed to get kafka broker metadata: %w", err)
	}

	fmt.Printf("📡 Kafka cluster has %d broker(s)\n", len(metadata.Brokers))
	for _, b := range metadata.Brokers {
		fmt.Printf("   broker %d: %s:%d\n", b.NodeID, b.Host, b.Port)
	}

	return nil
}

func EnsureTopics(ctx context.Context, brokers []string, topics []string, opts ...kgo.Opt) error {
	baseOpts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
	}
	baseOpts = append(baseOpts, opts...)

	client, err := kgo.NewClient(baseOpts...)
	if err != nil {
		return fmt.Errorf("failed to create kafka admin client: %w", err)
	}
	defer client.Close()

	admin := kadm.NewClient(client)

	ensureCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	responses, err := admin.CreateTopics(ensureCtx, 1, -1, nil, topics...)
	if err != nil {
		return fmt.Errorf("failed to create topics: %w", err)
	}

	for _, r := range responses.Sorted() {
		if r.Err != nil {
			fmt.Printf("⚠️  Topic %q: %v\n", r.Topic, r.Err)
		} else {
			fmt.Printf("✅ Topic %q ready\n", r.Topic)
		}
	}

	return nil
}
