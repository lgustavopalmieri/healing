package listener

import "errors"

const (
	SpecialistCreatedEventName = "specialist.created"
	WelcomeEmailTemplate       = "welcome"
	DefaultLocale              = "pt-BR"

	ErrInvalidEventPayloadMessage   = "Invalid event payload type"
	ErrUnmarshalEventPayloadMessage = "Failed to unmarshal event payload"
	ErrSendEmailMessage             = "Failed to send welcome email"
)

var (
	ErrInvalidEventPayload   = errors.New(ErrInvalidEventPayloadMessage)
	ErrUnmarshalEventPayload = errors.New(ErrUnmarshalEventPayloadMessage)
	ErrSendEmail             = errors.New(ErrSendEmailMessage)
)
