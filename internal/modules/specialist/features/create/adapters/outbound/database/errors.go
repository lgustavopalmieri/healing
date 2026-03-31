package database

var (
	FailedToSaveErr            = "failed to save specialist: %w"
	FailedToCheckUniquenessErr = "failed to check uniqueness: %w"

	IdAlreadyExistsErr      = "specialist with ID %s already exists"
	EmailAlreadyExistsErr   = "specialist with email %s already exists"
	LicenseAlreadyExistsErr = "specialist with license number %s already exists"
)
