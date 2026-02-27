package kafka

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	kafkalib "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application/listener"
	dbrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/infra/database"
	esrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/infra/elasticsearch"
	platformkafka "github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
)

type ManagerDependencies struct {
	DB                 *sql.DB
	ESClient           *elasticsearch.Client
	ESIndexSpecialists string
	Tracer             observability.Tracer
	Logger             observability.Logger
	KafkaConfig        *kafkalib.ConfigMap
}

func NewSpecialistUpdateKafkaManager(ctx context.Context, deps ManagerDependencies) error {
	repository := dbrepo.NewSpecialistFindByIDRepository(deps.DB)
	esProjection := esrepo.NewRepository(deps.ESClient, deps.ESIndexSpecialists, deps.Logger)

	projections := []listener.SpecialistReadProjectionInterface{
		esProjection,
	}

	specialistUpdatedListener := listener.NewSpecialistUpdatedListener(
		repository,
		projections,
		deps.Tracer,
		deps.Logger,
	)

	manager := event.NewListenerManager()
	manager.Register(listener.SpecialistUpdatedEventName, specialistUpdatedListener)

	consumer, err := platformkafka.NewKafkaConsumer(deps.KafkaConfig, manager)
	if err != nil {
		return fmt.Errorf("failed to create specialist update kafka consumer: %w", err)
	}

	go consumer.Start(ctx)

	log.Println("✅ Specialist update kafka consumer started")

	return nil
}
