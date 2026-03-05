package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type UpdateSpecialistUseCase struct {
	repository     SpecialistUpdateRepositoryInterface
	eventPublisher event.EventDispatcher
	tracer         observability.Tracer
	logger         observability.Logger
}

func NewUpdateSpecialistUseCase(
	repository SpecialistUpdateRepositoryInterface,
	eventPublisher event.EventDispatcher,
	tracer observability.Tracer,
	logger observability.Logger,
) *UpdateSpecialistUseCase {
	return &UpdateSpecialistUseCase{
		repository:     repository,
		eventPublisher: eventPublisher,
		tracer:         tracer,
		logger:         logger,
	}
}
