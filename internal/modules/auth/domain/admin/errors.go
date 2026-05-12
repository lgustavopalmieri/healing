package admin

import "errors"

var (
	ErrInvalidStatusTransition = errors.New("invalid admin status transition")
	ErrInvalidSubRole          = errors.New("invalid admin sub-role")
)
