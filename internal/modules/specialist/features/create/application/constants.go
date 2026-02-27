package application

import "errors"

const (
	// Span names
	CreateSpecialistSpanName = "CreateSpecialistCommand.Execute"

	// Event names
	SpecialistCreatedEventName = "specialist.created"

	// Success messages
	StartingSpecialistCreationMessage = "Starting specialist creation"
	SpecialistCreatedSuccessMessage   = "Specialist created successfully"

	// Error messages
	ErrUniquenessValidationMessage = "Failed to validate uniqueness constraints"
	ErrSaveSpecialistMessage       = "Failed to save specialist"
	ErrEventPublishMessage         = "Failed to publish specialist created event"
)

// Application errors
var (
	ErrUniquenessValidation = errors.New(ErrUniquenessValidationMessage)
	ErrSaveSpecialist       = errors.New(ErrSaveSpecialistMessage)
	ErrEventPublish         = errors.New(ErrEventPublishMessage)
)
