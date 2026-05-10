package database

var (
	FailedToFindCredentialErr  = "failed to find credential: %w"
	FailedToSaveCredentialErr  = "failed to save credential: %w"
	CredentialAlreadyExistsErr = "credential already exists for email/provider/role combination"
)
