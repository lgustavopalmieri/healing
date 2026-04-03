package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type CreateSpecialistUseCase struct {
	repository     SpecialistCreateRepositoryInterface
	eventPublisher event.EventDispatcher
	logger         observability.Logger
}

func NewCreateSpecialistUseCase(
	repository SpecialistCreateRepositoryInterface,
	eventPublisher event.EventDispatcher,
	logger observability.Logger,
) *CreateSpecialistUseCase {
	return &CreateSpecialistUseCase{
		repository:     repository,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}
