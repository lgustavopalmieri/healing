package bootstrap

import (
	"fmt"
	"log"

	kafkalib "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
)

func InitKafkaProducer(cfg *config.Config) (*kafka.KafkaProducer, error) {
	log.Printf("📨 Connecting to Kafka broker (%s)...", cfg.Kafka.BootstrapServers)
	
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

	log.Println("✅ Kafka producer initialized successfully")

	return producer, nil
}
