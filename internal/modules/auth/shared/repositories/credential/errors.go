package credential

const (
	FailedToFindCredentialErr   = "failed to find credential: %w"
	FailedToSaveCredentialErr   = "failed to save credential: %w"
	FailedToUpdateCredentialErr = "failed to update credential: %w"
	FailedToSaveSessionErr      = "failed to save session: %w"
	CredentialAlreadyExistsErr  = "credential already exists for email/provider/role combination"
)
