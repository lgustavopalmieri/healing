package listener

const (
	SpecialistUpdatedEventName = "specialist.updated"

	SpecialistUpdatedListenerSpanName = "SpecialistUpdatedListener.Handle"

	StartingSpecialistUpdatedMessage          = "Starting specialist updated event processing"
	SpecialistUpdatedProcessedSuccessMessage  = "Specialist updated event processed successfully"
	ErrUnmarshalSpecialistUpdatedEventMessage = "Failed to unmarshal specialist updated event payload"
	ErrFindSpecialistByIDMessage              = "Failed to find specialist by ID in database"
	ErrUpdateProjectionMessage                = "Failed to update read projection"
)
