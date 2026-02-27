package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	kafkalib "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	specialistupdatekafka "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/infra/event_listeners/kafka"
)

type ConsumerDependencies struct {
	DB                 *sql.DB
	ESClient           *elasticsearch.Client
	ESIndexSpecialists string
	Tracer             observability.Tracer
	Logger             observability.Logger
	Config             *config.Config
}

func InitKafkaConsumers(ctx context.Context, deps ConsumerDependencies) error {
	log.Println("📨 Starting Kafka consumers...")

	kafkaConfig := &kafkalib.ConfigMap{
		"bootstrap.servers":  deps.Config.Kafka.BootstrapServers,
		"group.id":           "specialist-update-consumer-group",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	err := specialistupdatekafka.NewSpecialistUpdateKafkaManager(ctx, specialistupdatekafka.ManagerDependencies{
		DB:                 deps.DB,
		ESClient:           deps.ESClient,
		ESIndexSpecialists: deps.ESIndexSpecialists,
		Tracer:             deps.Tracer,
		Logger:             deps.Logger,
		KafkaConfig:        kafkaConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize specialist update kafka consumer: %w", err)
	}

	log.Println("✅ Kafka consumers initialized successfully")
	return nil
}
