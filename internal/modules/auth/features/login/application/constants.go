package application

import "errors"

const (
	FailedToFindCredentialMessage = "Failed to find credential for login"
	FailedToIssueTokenPairMessage = "Failed to issue token pair"
	FailedToPersistSessionMessage = "Failed to persist session"
	FailedToCacheRefreshMessage   = "Failed to cache refresh token"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCredentialLocked   = errors.New("credential locked")
	ErrIssueTokens        = errors.New(FailedToIssueTokenPairMessage)
	ErrPersistSession     = errors.New(FailedToPersistSessionMessage)
	ErrCacheRefreshToken  = errors.New(FailedToCacheRefreshMessage)
)
