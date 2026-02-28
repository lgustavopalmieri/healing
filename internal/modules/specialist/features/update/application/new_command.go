package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type UpdateSpecialistCommand struct {
	repository     SpecialistUpdateRepositoryInterface
	eventPublisher event.EventDispatcher
	tracer         observability.Tracer
	logger         observability.Logger
}

func NewUpdateSpecialistCommand(
	repository SpecialistUpdateRepositoryInterface,
	eventPublisher event.EventDispatcher,
	tracer observability.Tracer,
	logger observability.Logger,
) *UpdateSpecialistCommand {
	return &UpdateSpecialistCommand{
		repository:     repository,
		eventPublisher: eventPublisher,
		tracer:         tracer,
		logger:         logger,
	}
}
