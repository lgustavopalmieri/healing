package password

import "errors"

var (
	ErrTooShort             = errors.New("password too short")
	ErrMissingRequiredChars = errors.New("password must contain letters and numbers")
)
