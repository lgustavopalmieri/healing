package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type CreateSpecialistCommand struct {
	repository      SpecialistCreateRepositoryInterface
	externalGateway SpecialistCreateExternalGatewayInterface
	eventPublisher  event.EventDispatcher
	tracer          observability.Tracer
	logger          observability.Logger
}

func NewCreateSpecialistCommand(
	repository SpecialistCreateRepositoryInterface,
	externalGateway SpecialistCreateExternalGatewayInterface,
	eventPublisher event.EventDispatcher,
	tracer observability.Tracer,
	logger observability.Logger,
) *CreateSpecialistCommand {
	return &CreateSpecialistCommand{
		repository:      repository,
		externalGateway: externalGateway,
		eventPublisher:  eventPublisher,
		tracer:          tracer,
		logger:          logger,
	}
}
