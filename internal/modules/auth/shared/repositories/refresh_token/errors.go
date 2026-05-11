package refreshtoken

const (
	FailedToSaveRefreshTokenErr   = "failed to save refresh token in cache: %w"
	FailedToFindRefreshTokenErr   = "failed to find refresh token in cache: %w"
	FailedToDeleteRefreshTokenErr = "failed to delete refresh token from cache: %w"
	ErrInvalidRefreshTokenTTL     = "invalid refresh token TTL: %d"
)
