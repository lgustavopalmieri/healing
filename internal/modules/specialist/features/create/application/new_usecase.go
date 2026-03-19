package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

type CreateSpecialistUseCase struct {
	repository     SpecialistCreateRepositoryInterface
	eventPublisher event.EventDispatcher
}

func NewCreateSpecialistUseCase(
	repository SpecialistCreateRepositoryInterface,
	eventPublisher event.EventDispatcher,
) *CreateSpecialistUseCase {
	return &CreateSpecialistUseCase{
		repository:     repository,
		eventPublisher: eventPublisher,
	}
}
