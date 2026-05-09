package errors

import "errors"

var (
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token expired")
	ErrBlacklistedToken   = errors.New("token revoked")
	ErrForbidden          = errors.New("forbidden")
	ErrForbiddenNotOwner  = errors.New("forbidden: not owner")
	ErrForbiddenWrongRole = errors.New("forbidden: wrong role")
	ErrNoClaims           = errors.New("no claims in context")
	ErrInvalidClaims      = errors.New("invalid claims")
)
