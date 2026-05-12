package application

import "errors"

var (
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrDeleteOldRefresh    = errors.New("failed to delete old refresh token")
	ErrIssueNewTokens      = errors.New("failed to issue new token pair")
	ErrCacheNewRefresh     = errors.New("failed to cache new refresh token")
)
