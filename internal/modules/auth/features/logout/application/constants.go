package application

import "errors"

var (
	ErrDeleteRefreshToken   = errors.New("failed to delete refresh token")
	ErrBlacklistAccessToken = errors.New("failed to blacklist access token")
)
