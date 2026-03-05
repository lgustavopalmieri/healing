package application

import "errors"

const (
	UpdateSpecialistSpanName = "UpdateSpecialistUseCase.Execute"

	SpecialistUpdatedEventName = "specialist.updated"

	StartingSpecialistUpdateMessage = "Starting specialist update"
	SpecialistUpdatedSuccessMessage = "Specialist updated successfully"

	ErrSpecialistNotFoundMessage = "Specialist not found"
	ErrUpdateSpecialistMessage   = "Failed to update specialist"
	ErrEventPublishMessage       = "Failed to publish specialist updated event"
)

var (
	ErrSpecialistNotFound = errors.New(ErrSpecialistNotFoundMessage)
	ErrUpdateSpecialist   = errors.New(ErrUpdateSpecialistMessage)
	ErrEventPublish       = errors.New(ErrEventPublishMessage)
)
