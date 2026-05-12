package application

import "errors"

const (
	FailedToConsumeSingleUseTokenMessage = "Failed to consume single-use set-password token"
	FailedToFindCredentialMessage        = "Failed to find credential"
	FailedToHashPasswordMessage          = "Failed to hash password"
	FailedToActivateCredentialMessage    = "Failed to activate credential"
	FailedToPersistCredentialMessage     = "Failed to persist credential"
	FailedToIssueTokenPairMessage        = "Failed to issue access and refresh tokens"
	FailedToPersistSessionMessage        = "Failed to persist session"
	FailedToCacheRefreshTokenMessage     = "Failed to cache refresh token"
	FailedToRecordAuditMessage           = "Failed to record audit log"
	FailedToPublishActivatedEventMessage = "Failed to publish auth.credential.activated event"
)

var (
	ErrInvalidSetPasswordToken    = errors.New("invalid or expired set-password token")
	ErrSingleUseTokenAlreadyUsed  = errors.New("set-password token already used")
	ErrCredentialNotFound         = errors.New("credential not found for token subject")
	ErrCredentialNotPending       = errors.New("credential is not in pending state")
	ErrFailedToHashPassword       = errors.New(FailedToHashPasswordMessage)
	ErrFailedToPersistCredential  = errors.New(FailedToPersistCredentialMessage)
	ErrFailedToIssueTokenPair     = errors.New(FailedToIssueTokenPairMessage)
	ErrFailedToPersistSession     = errors.New(FailedToPersistSessionMessage)
	ErrFailedToCacheRefreshToken  = errors.New(FailedToCacheRefreshTokenMessage)
	ErrFailedToConsumeSingleUse   = errors.New(FailedToConsumeSingleUseTokenMessage)
	ErrFailedToFindCredential     = errors.New(FailedToFindCredentialMessage)
	ErrFailedToActivateCredential = errors.New(FailedToActivateCredentialMessage)
)
