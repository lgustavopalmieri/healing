package application

import "errors"

const (
	SpecialistCreatedEventName = "specialist.created"

	ErrUniquenessValidationMessage = "Failed to validate uniqueness constraints"
	ErrSaveSpecialistMessage       = "Failed to save specialist"
	ErrEventPublishMessage         = "Failed to publish specialist created event"
)

var (
	ErrUniquenessValidation = errors.New(ErrUniquenessValidationMessage)
	ErrSaveSpecialist       = errors.New(ErrSaveSpecialistMessage)
	ErrEventPublish         = errors.New(ErrEventPublishMessage)
)
