package database

var (
	FailedToFindByIDErr     = "failed to find specialist by ID: %w"
	SpecialistNotFoundErr   = "specialist with ID %s not found"
	FailedToUpdateStatusErr = "failed to update specialist status: %w"
	UpdateStatusNotFoundErr = "specialist with ID %s not found for status update"
)
