package bootstrap

import (
	"context"
	"database/sql"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	validatelicensesqs "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/adapters/inbound/sqs"
	updateapp "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
	updatedatarepossqs "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/adapters/inbound/sqs"
	platformOS "github.com/lgustavopalmieri/healing-specialist/internal/platform/opensearch"
	opensearchindexes "github.com/lgustavopalmieri/healing-specialist/internal/platform/opensearch/indexes"
)

type SQSConsumerDependencies struct {
	DB             *sql.DB
	OSFactory      *platformOS.Factory
	EventPublisher event.EventDispatcher
	SQS            *SQSResources
	Config         *config.Config
}

func InitSQSConsumers(ctx context.Context, deps SQSConsumerDependencies) {
	log.Println("Starting SQS consumers...")

	createdQueueURL := deps.SQS.QueueURLs[application.SpecialistCreatedEventName]
	updatedQueueURL := deps.SQS.QueueURLs[updateapp.SpecialistUpdatedEventName]

	validatelicensesqs.NewValidateLicenseSQSManager(ctx, validatelicensesqs.ManagerDependencies{
		DB:              deps.DB,
		EventDispatcher: deps.EventPublisher,
		LicenseBaseURL:  deps.Config.External.LicenseBaseURL,
		SQSClient:       deps.SQS.Client,
		QueueURL:        createdQueueURL,
	})

	updatedatarepossqs.NewUpdateDataRepositoriesSQSManager(ctx, updatedatarepossqs.ManagerDependencies{
		DB:        deps.DB,
		OSClient:  deps.OSFactory.Client,
		IndexName: deps.OSFactory.IndexName(opensearchindexes.SpecialistsIndex),
		SQSClient: deps.SQS.Client,
		QueueURL:  updatedQueueURL,
	})

	log.Println("SQS consumers initialized successfully")
}
