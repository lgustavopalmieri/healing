package sqs

import (
	"context"
	"database/sql"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/adapters/outbound/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/adapters/outbound/opensearch"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/listener"
	platformsqs "github.com/lgustavopalmieri/healing-specialist/internal/platform/sqs"
)

type ManagerDependencies struct {
	DB        *sql.DB
	OSClient  *opensearchapi.Client
	IndexName string
	SQSClient *awssqs.Client
	QueueURL  string
}

func NewUpdateDataRepositoriesSQSManager(ctx context.Context, deps ManagerDependencies) {
	sourceRepo := database.NewSourceRepository(deps.DB)

	osRepo := opensearch.NewRepository(deps.OSClient, deps.IndexName)

	dataRepositories := []listener.DataRepository{
		osRepo,
	}

	handler := listener.NewUpdateDataRepositoriesHandler(
		sourceRepo,
		dataRepositories,
	)

	consumer := platformsqs.NewSQSConsumer(
		deps.SQSClient,
		deps.QueueURL,
		application.SpecialistUpdatedEventName,
		handler,
	)

	go consumer.Start(ctx)
}
