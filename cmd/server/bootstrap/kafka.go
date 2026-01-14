package bootstrap

import (
	"fmt"

	kafkalib "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
)

func InitKafkaProducer(cfg *config.Config) (*kafka.KafkaProducer, error) {
	kafkaConfig := &kafkalib.ConfigMap{
		"bootstrap.servers": cfg.Kafka.BootstrapServers,
	}

	if cfg.Kafka.AutoOffsetReset != "" {
		kafkaConfig.SetKey("auto.offset.reset", cfg.Kafka.AutoOffsetReset)
	}

	producer, err := kafka.NewKafkaProducer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kafka producer: %w", err)
	}

	return producer, nil
}
