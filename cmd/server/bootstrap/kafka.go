package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	createapp "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	updateapp "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
	updatedatareposlistener "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/listener"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
)

func InitKafkaProducer(cfg *config.Config) (*kafka.KafkaProducer, error) {
	log.Printf("📨 Connecting to Kafka broker (%s)...", cfg.Kafka.BootstrapServers)

	brokers := []string{cfg.Kafka.BootstrapServers}

	authOpts, err := kafka.AuthOpts(kafka.AuthConfig{
		SASLMechanism: cfg.Kafka.SASLMechanism,
		SASLUsername:  cfg.Kafka.SASLUsername,
		SASLPassword:  cfg.Kafka.SASLPassword,
		UseTLS:        cfg.Kafka.UseTLS,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build kafka auth opts: %w", err)
	}

	log.Println("🔍 Running Kafka health check...")
	if err := kafka.HealthCheck(context.Background(), brokers, authOpts...); err != nil {
		return nil, fmt.Errorf("kafka health check failed: %w", err)
	}

	topics := []string{
		createapp.SpecialistCreatedEventName,
		updateapp.SpecialistUpdatedEventName,
		updatedatareposlistener.UpdateDataRepositoriesDLQEventName,
	}

	log.Println("📋 Ensuring Kafka topics exist...")
	if err := kafka.EnsureTopics(context.Background(), brokers, topics, authOpts...); err != nil {
		return nil, fmt.Errorf("failed to ensure kafka topics: %w", err)
	}

	producer, err := kafka.NewKafkaProducer(brokers, authOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kafka producer: %w", err)
	}

	log.Println("✅ Kafka producer initialized successfully")

	return producer, nil
}
