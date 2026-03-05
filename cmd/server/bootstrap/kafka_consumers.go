package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	validatelicensekafka "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/infra/kafka"
)

type ConsumerDependencies struct {
	DB                 *sql.DB
	ESClient           *elasticsearch.Client
	ESIndexSpecialists string
	Tracer             observability.Tracer
	Logger             observability.Logger
	EventPublisher     event.EventDispatcher
	Config             *config.Config
}

func InitKafkaConsumers(ctx context.Context, deps ConsumerDependencies) error {
	log.Println("📨 Starting Kafka consumers...")

	err := validatelicensekafka.NewValidateLicenseKafkaManager(ctx, validatelicensekafka.ManagerDependencies{
		DB:               deps.DB,
		Tracer:           deps.Tracer,
		Logger:           deps.Logger,
		EventDispatcher:  deps.EventPublisher,
		LicenseBaseURL:   deps.Config.External.LicenseBaseURL,
		BootstrapServers: deps.Config.Kafka.BootstrapServers,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize validate license kafka consumer: %w", err)
	}

	log.Println("✅ Kafka consumers initialized successfully")
	return nil
}
