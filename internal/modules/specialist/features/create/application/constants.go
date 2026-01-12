package application

import "errors"

const (
	// Timeouts
	ExternalValidationTimeoutMs = 800 // milliseconds

	// Span names
	CreateSpecialistSpanName = "CreateSpecialistCommand.Execute"

	// Event names
	SpecialistCreatedEventName = "specialist.created"

	// Success messages
	StartingSpecialistCreationMessage = "Starting specialist creation"
	SpecialistCreatedSuccessMessage   = "Specialist created successfully"

	// Error messages
	ErrUniquenessValidationMessage  = "Failed to validate uniqueness constraints"
	ErrSaveSpecialistMessage        = "Failed to save specialist"
	ErrLicenseValidationMessage     = "Failed to validate license number"
	ErrEventPublishMessage          = "Failed to publish specialist created event"
	ErrExternalValidationTimeoutMsg = "License Number validation timeout. Try again."

	// Warning messages
	InvalidLicenseNumberMessage = "Invalid license number"
)

// Application errors
var (
	ErrUniquenessValidation      = errors.New(ErrUniquenessValidationMessage)
	ErrSaveSpecialist            = errors.New(ErrSaveSpecialistMessage)
	ErrLicenseValidation         = errors.New(ErrLicenseValidationMessage)
	ErrEventPublish              = errors.New(ErrEventPublishMessage)
	ErrExternalValidationTimeout = errors.New(ErrExternalValidationTimeoutMsg)
)
