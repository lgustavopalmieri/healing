package kafka

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	vlApplication "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/infra/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/infra/external"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/infra/listener"
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

	command := vlApplication.NewValidateLicenseCommand(
		repository,
		gateway,
		deps.EventDispatcher,
		deps.Tracer,
		deps.Logger,
	)

	validateLicenseListener := listener.NewValidateLicenseListener(command)

	manager := event.NewListenerManager()
	manager.Register(application.SpecialistCreatedEventName, validateLicenseListener)

	config := &libkafka.ConfigMap{
		"bootstrap.servers":  deps.BootstrapServers,
		"group.id":           "specialist-validate-license-consumer-group",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	consumer, err := platformkafka.NewKafkaConsumer(config, manager)
	if err != nil {
		return fmt.Errorf("failed to create validate license kafka consumer: %w", err)
	}

	go consumer.Start(ctx)

	return nil
}
