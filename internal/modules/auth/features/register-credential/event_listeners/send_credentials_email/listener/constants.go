package listener

import "errors"

const (
	AuthCredentialPendingEventName = "auth.credential.pending"
	SetPasswordEmailTemplate       = "set-password"
	DefaultLocale                  = "pt-BR"

	ErrInvalidEventPayloadMessage   = "Invalid event payload type"
	ErrUnmarshalEventPayloadMessage = "Failed to unmarshal auth.credential.pending payload"
	ErrSendEmailMessage             = "Failed to send set-password email"
)

var (
	ErrInvalidEventPayload   = errors.New(ErrInvalidEventPayloadMessage)
	ErrUnmarshalEventPayload = errors.New(ErrUnmarshalEventPayloadMessage)
	ErrSendEmail             = errors.New(ErrSendEmailMessage)
)
