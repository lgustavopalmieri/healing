package database

var (
	FailedToSaveErr         = "failed to save specialist: %w"
	FailedToCheckIdErr      = "failed to check ID uniqueness: %w"
	FailedToCheckEmailErr   = "failed to check email uniqueness: %w"
	FailedToCheckLicenseErr = "failed to check license number uniqueness: %w"

	IdAlreadyExistsErr      = "specialist with ID %s already exists"
	EmailAlreadyExistsErr   = "specialist with email %s already exists"
	LicenseAlreadyExistsErr = "specialist with license number %s already exists"
)
