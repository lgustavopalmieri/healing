package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

type UpdateSpecialistUseCase struct {
	repository     SpecialistUpdateRepositoryInterface
	eventPublisher event.EventDispatcher
}

func NewUpdateSpecialistUseCase(
	repository SpecialistUpdateRepositoryInterface,
	eventPublisher event.EventDispatcher,
) *UpdateSpecialistUseCase {
	return &UpdateSpecialistUseCase{
		repository:     repository,
		eventPublisher: eventPublisher,
	}
}
