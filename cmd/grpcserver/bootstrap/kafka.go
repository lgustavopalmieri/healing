package bootstrap

import (
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
)

func InitKafkaProducer(cfg *config.Config) (*kafka.KafkaProducer, error) {
	log.Printf("📨 Connecting to Kafka broker (%s)...", cfg.Kafka.BootstrapServers)

	brokers := []string{cfg.Kafka.BootstrapServers}

	producer, err := kafka.NewKafkaProducer(brokers)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kafka producer: %w", err)
	}

	log.Println("✅ Kafka producer initialized successfully")

	return producer, nil
}
