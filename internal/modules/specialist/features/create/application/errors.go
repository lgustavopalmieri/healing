package application

import "errors"

// Constants
const (
	// Span names
	CreateSpecialistSpanName = "CreateSpecialistCommand.Execute"

	// Event names
	SpecialistCreatedEventName = "specialist.created"

	// Success messages
	StartingSpecialistCreationMessage = "Starting specialist creation"
	SpecialistCreatedSuccessMessage   = "Specialist created successfully"

	// Error messages
	ErrEmailCheckMessage        = "Failed to check email existence"
	ErrLicenseCheckMessage      = "Failed to check license number existence"
	ErrIDCheckMessage           = "Failed to check ID existence"
	ErrSaveSpecialistMessage    = "Failed to save specialist"
	ErrLicenseValidationMessage = "Failed to validate license number"
	ErrEventPublishMessage      = "Failed to publish specialist created event"

	// Warning messages
	EmailAlreadyExistsMessage   = "Email already exists"
	LicenseAlreadyExistsMessage = "License number already exists"
	InvalidLicenseNumberMessage = "Invalid license number"
	IDAlreadyExistsMessage      = "Generated ID already exists"
)

// Application errors
var (
	ErrEmailCheck        = errors.New(ErrEmailCheckMessage)
	ErrLicenseCheck      = errors.New(ErrLicenseCheckMessage)
	ErrIDCheck           = errors.New(ErrIDCheckMessage)
	ErrSaveSpecialist    = errors.New(ErrSaveSpecialistMessage)
	ErrLicenseValidation = errors.New(ErrLicenseValidationMessage)
	ErrEventPublish      = errors.New(ErrEventPublishMessage)
)
