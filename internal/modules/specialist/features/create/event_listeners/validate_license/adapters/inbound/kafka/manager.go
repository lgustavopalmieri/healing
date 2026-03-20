package kafka

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/adapters/outbound/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/adapters/outbound/external"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/listener"
	platformkafka "github.com/lgustavopalmieri/healing-specialist/internal/platform/kafka"
)

type ManagerDependencies struct {
	DB               *sql.DB
	EventDispatcher  event.EventDispatcher
	LicenseBaseURL   string
	BootstrapServers string
	KafkaAuthConfig  platformkafka.AuthConfig
}

func NewValidateLicenseKafkaManager(ctx context.Context, deps ManagerDependencies) error {
	repository := database.NewValidateLicenseRepository(deps.DB)
	gateway := external.NewLicenseGateway(deps.LicenseBaseURL, &http.Client{})

	handler := listener.NewValidateLicenseHandler(
		repository,
		gateway,
		deps.EventDispatcher,
	)

	manager := event.NewListenerManager()
	manager.Register(application.SpecialistCreatedEventName, handler)

	authOpts, err := platformkafka.AuthOpts(deps.KafkaAuthConfig)
	if err != nil {
		return err
	}

	consumer, err := platformkafka.NewKafkaConsumer(
		[]string{deps.BootstrapServers},
		"specialist-validate-license-consumer-group",
		manager,
		authOpts...,
	)
	if err != nil {
		return err
	}

	go consumer.Start(ctx)

	return nil
}
