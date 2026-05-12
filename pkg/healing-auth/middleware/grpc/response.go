package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
)

func MapError(err error) error {
	switch {
	case errors.Is(err, autherrors.ErrUnauthenticated),
		errors.Is(err, autherrors.ErrInvalidToken),
		errors.Is(err, autherrors.ErrExpiredToken),
		errors.Is(err, autherrors.ErrBlacklistedToken),
		errors.Is(err, autherrors.ErrNoClaims),
		errors.Is(err, autherrors.ErrInvalidClaims):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, autherrors.ErrForbidden),
		errors.Is(err, autherrors.ErrForbiddenNotOwner),
		errors.Is(err, autherrors.ErrForbiddenWrongRole):
		return status.Error(codes.PermissionDenied, err.Error())
	}
	return status.Error(codes.Unauthenticated, err.Error())
}
