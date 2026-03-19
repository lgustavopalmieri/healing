package kafka

import (
	"context"
	"database/sql"

	goelasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/adapters/outbound/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/adapters/outbound/elasticsearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/listener"
	platformkafka "github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
)

type ManagerDependencies struct {
	DB                 *sql.DB
	ESClient           *goelasticsearch.Client
	ESIndexSpecialists string
	EventDispatcher    event.EventDispatcher
	BootstrapServers   string
}

func NewUpdateDataRepositoriesKafkaManager(ctx context.Context, deps ManagerDependencies) error {
	sourceRepo := database.NewSourceRepository(deps.DB)

	esRepo := elasticsearch.NewRepository(deps.ESClient, deps.ESIndexSpecialists, deps.EventDispatcher)

	dataRepositories := []listener.DataRepository{
		esRepo,
	}

	handler := listener.NewUpdateDataRepositoriesHandler(
		sourceRepo,
		dataRepositories,
	)

	manager := event.NewListenerManager()
	manager.Register(application.SpecialistUpdatedEventName, handler)

	consumer, err := platformkafka.NewKafkaConsumer(
		[]string{deps.BootstrapServers},
		"specialist-update-data-repositories-consumer-group",
		manager,
	)
	if err != nil {
		return err
	}

	go consumer.Start(ctx)

	return nil
}
