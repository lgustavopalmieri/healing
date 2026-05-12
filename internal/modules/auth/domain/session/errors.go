package session

import "errors"

var (
	ErrAlreadyRevoked = errors.New("session already revoked")
)
