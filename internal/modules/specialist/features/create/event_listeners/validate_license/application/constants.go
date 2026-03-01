package application

import "errors"

const (
	ValidateLicenseSpanName = "ValidateLicenseCommand.Execute"

	SpecialistUpdatedEventName = "specialist.updated"

	StartingLicenseValidationMessage = "Starting license validation"
	LicenseValidatedSuccessMessage   = "License validated successfully"
	SpecialistStatusUpdatedMessage   = "Specialist status updated to active"

	ErrSpecialistNotFoundMessage    = "Specialist not found for license validation"
	ErrLicenseValidationMessage     = "Failed to validate license with external service"
	ErrInvalidLicenseMessage        = "License is not valid"
	ErrUpdateStatusMessage          = "Failed to update specialist status"
	ErrEventPublishMessage          = "Failed to publish specialist updated event"
	ErrUnmarshalEventPayloadMessage = "Failed to unmarshal event payload"
)

var (
	ErrSpecialistNotFound    = errors.New(ErrSpecialistNotFoundMessage)
	ErrLicenseValidation     = errors.New(ErrLicenseValidationMessage)
	ErrInvalidLicense        = errors.New(ErrInvalidLicenseMessage)
	ErrUpdateStatus          = errors.New(ErrUpdateStatusMessage)
	ErrEventPublish          = errors.New(ErrEventPublishMessage)
	ErrUnmarshalEventPayload = errors.New(ErrUnmarshalEventPayloadMessage)
)
