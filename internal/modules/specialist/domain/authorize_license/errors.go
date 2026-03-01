package authorizelicense

import "errors"

var (
	ErrInvalidStatusTransition = errors.New("specialist must be in pending status to authorize license")
)
