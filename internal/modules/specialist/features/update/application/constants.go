package application

import "errors"

const (
	SpecialistUpdatedEventName = "specialist.updated"

	ErrSpecialistNotFoundMessage = "Specialist not found"
	ErrUpdateSpecialistMessage   = "Failed to update specialist"
	ErrForbiddenNotOwnerMessage  = "Forbidden: not the owner of this resource"
)

var (
	ErrSpecialistNotFound = errors.New(ErrSpecialistNotFoundMessage)
	ErrUpdateSpecialist   = errors.New(ErrUpdateSpecialistMessage)
	ErrForbiddenNotOwner  = errors.New(ErrForbiddenNotOwnerMessage)
)
