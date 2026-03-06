package command

import "errors"

const (
	UpdateDataRepositoriesSpanName = "UpdateDataRepositoriesHandler.Handle"

	UpdateDataRepositoriesDLQEventName = "specialist.updated.dlq"

	StartingDataRepositoriesUpdateMessage = "Starting data repositories update"
	DataRepositoriesUpdatedSuccessMessage = "Data repositories updated successfully"
	RepositoryUpdateStartedMessage        = "Starting repository update"
	RepositoryUpdateSucceededMessage      = "Repository update succeeded"
	RepositoryUpdateFailedMessage         = "Repository update failed after retries, publishing to DLQ"
	DLQPublishFailedMessage               = "Failed to publish DLQ event"

	ErrSpecialistNotFoundMessage     = "Specialist not found for data repositories update"
	ErrUnmarshalEventPayloadMessage  = "Failed to unmarshal event payload"
	ErrUpdateDataRepositoriesMessage = "Failed to update one or more data repositories"
)

var (
	ErrSpecialistNotFound     = errors.New(ErrSpecialistNotFoundMessage)
	ErrUnmarshalEventPayload  = errors.New(ErrUnmarshalEventPayloadMessage)
	ErrUpdateDataRepositories = errors.New(ErrUpdateDataRepositoriesMessage)
)
