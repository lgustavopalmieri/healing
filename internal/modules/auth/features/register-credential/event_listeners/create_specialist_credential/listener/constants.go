package listener

import "errors"

const (
	SpecialistCreatedEventName = "specialist.created"

	ErrInvalidEventPayloadMessage   = "Invalid event payload type"
	ErrUnmarshalEventPayloadMessage = "Failed to unmarshal event payload"
	ErrFindCredentialMessage        = "Failed to find existing credential"
	ErrSaveCredentialMessage        = "Failed to save credential"
	ErrGenerateSetPasswordMessage   = "Failed to generate set-password token"
	ErrPublishCredentialPending     = "Failed to publish auth.credential.pending event"
)

var (
	ErrInvalidEventPayload   = errors.New(ErrInvalidEventPayloadMessage)
	ErrUnmarshalEventPayload = errors.New(ErrUnmarshalEventPayloadMessage)
	ErrFindCredential        = errors.New(ErrFindCredentialMessage)
	ErrSaveCredential        = errors.New(ErrSaveCredentialMessage)
	ErrGenerateSetPassword   = errors.New(ErrGenerateSetPasswordMessage)
	ErrPublishPending        = errors.New(ErrPublishCredentialPending)
)
