package database

var (
	FailedToFindByIDErr   = "failed to find specialist by ID: %w"
	SpecialistNotFoundErr = "specialist with ID %s not found"
	FailedToUpdateErr     = "failed to update specialist: %w"
	UpdateNotFoundErr     = "specialist with ID %s not found for update"
)
