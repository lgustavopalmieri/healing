package listener

import "errors"

const (
	ErrSpecialistNotFoundMessage     = "Specialist not found for data repositories update"
	ErrUnmarshalEventPayloadMessage  = "Failed to unmarshal event payload"
	ErrUpdateDataRepositoriesMessage = "Failed to update one or more data repositories"
)

var (
	ErrSpecialistNotFound     = errors.New(ErrSpecialistNotFoundMessage)
	ErrUnmarshalEventPayload  = errors.New(ErrUnmarshalEventPayloadMessage)
	ErrUpdateDataRepositories = errors.New(ErrUpdateDataRepositoriesMessage)
)
