package kafka

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/adapters/outbound/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/adapters/outbound/external"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/listener"
	platformkafka "github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
)

type ManagerDependencies struct {
	DB               *sql.DB
	Tracer           observability.Tracer
	Logger           observability.Logger
	EventDispatcher  event.EventDispatcher
	LicenseBaseURL   string
	BootstrapServers string
}

func NewValidateLicenseKafkaManager(ctx context.Context, deps ManagerDependencies) error {
	repository := database.NewValidateLicenseRepository(deps.DB)
	gateway := external.NewLicenseGateway(deps.LicenseBaseURL, &http.Client{})

	handler := listener.NewValidateLicenseHandler(
		repository,
		gateway,
		deps.EventDispatcher,
		deps.Tracer,
		deps.Logger,
	)

	manager := event.NewListenerManager()
	manager.Register(application.SpecialistCreatedEventName, handler)

	consumer, err := platformkafka.NewKafkaConsumer(
		[]string{deps.BootstrapServers},
		"specialist-validate-license-consumer-group",
		manager,
	)
	if err != nil {
		return err
	}

	go consumer.Start(ctx)

	return nil
}
