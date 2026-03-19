package application

import "errors"

const (
	SpecialistUpdatedEventName = "specialist.updated"

	ErrSpecialistNotFoundMessage = "Specialist not found"
	ErrUpdateSpecialistMessage   = "Failed to update specialist"
)

var (
	ErrSpecialistNotFound = errors.New(ErrSpecialistNotFoundMessage)
	ErrUpdateSpecialist   = errors.New(ErrUpdateSpecialistMessage)
)
