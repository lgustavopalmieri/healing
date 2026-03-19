package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	validatelicensekafka "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/adapters/inbound/kafka"
	updatedatareposkafka "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/adapters/inbound/kafka"
	platformES "github.com/lgustavopalmieri/healing-specialist/internal/platform/elasticsearch"
)

type ConsumerDependencies struct {
	DB             *sql.DB
	ESFactory      *platformES.Factory
	EventPublisher event.EventDispatcher
	Config         *config.Config
}

func InitKafkaConsumers(ctx context.Context, deps ConsumerDependencies) error {
	log.Println("📨 Starting Kafka consumers...")

	err := validatelicensekafka.NewValidateLicenseKafkaManager(ctx, validatelicensekafka.ManagerDependencies{
		DB:               deps.DB,
		EventDispatcher:  deps.EventPublisher,
		LicenseBaseURL:   deps.Config.External.LicenseBaseURL,
		BootstrapServers: deps.Config.Kafka.BootstrapServers,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize validate license kafka consumer: %w", err)
	}

	err = updatedatareposkafka.NewUpdateDataRepositoriesKafkaManager(ctx, updatedatareposkafka.ManagerDependencies{
		DB:                 deps.DB,
		ESClient:           deps.ESFactory.Client,
		ESIndexSpecialists: deps.ESFactory.Indexes.Specialists,
		EventDispatcher:    deps.EventPublisher,
		BootstrapServers:   deps.Config.Kafka.BootstrapServers,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize update data repositories kafka consumer: %w", err)
	}

	log.Println("✅ Kafka consumers initialized successfully")
	return nil
}
