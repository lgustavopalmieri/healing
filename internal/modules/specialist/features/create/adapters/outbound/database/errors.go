package database

var (
	FailedToBeginTxErr         = "failed to begin transaction: %w"
	FailedToCommitTxErr        = "failed to commit transaction: %w"
	FailedToSaveErr            = "failed to save specialist: %w"
	FailedToCheckUniquenessErr = "failed to check uniqueness: %w"

	IdAlreadyExistsErr      = "specialist with ID %s already exists"
	EmailAlreadyExistsErr   = "specialist with email %s already exists"
	LicenseAlreadyExistsErr = "specialist with license number %s already exists"
)
