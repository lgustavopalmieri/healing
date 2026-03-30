package sqs

import (
	"context"
	"database/sql"
	"net/http"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/adapters/outbound/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/adapters/outbound/external"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/listener"
	platformsqs "github.com/lgustavopalmieri/healing-specialist/internal/platform/sqs"
)

type ManagerDependencies struct {
	DB              *sql.DB
	EventDispatcher event.EventDispatcher
	LicenseBaseURL  string
	SQSClient       *awssqs.Client
	QueueURL        string
}

func NewValidateLicenseSQSManager(ctx context.Context, deps ManagerDependencies) {
	repository := database.NewValidateLicenseRepository(deps.DB)
	gateway := external.NewLicenseGateway(deps.LicenseBaseURL, &http.Client{})

	handler := listener.NewValidateLicenseHandler(
		repository,
		gateway,
		deps.EventDispatcher,
	)

	consumer := platformsqs.NewSQSConsumer(
		deps.SQSClient,
		deps.QueueURL,
		application.SpecialistCreatedEventName,
		handler,
	)

	go consumer.Start(ctx)
}
