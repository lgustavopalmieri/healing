package credential

import "errors"

var (
	ErrInvalidStatusTransition = errors.New("invalid credential status transition")
	ErrAlreadyActive           = errors.New("credential already active")
	ErrNotPending              = errors.New("credential not in pending state")
)
