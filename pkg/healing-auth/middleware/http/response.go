package http

import (
	"errors"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteError(c *gin.Context, err error) {
	status, code := mapError(err)
	c.AbortWithStatusJSON(status, ErrorResponse{
		Code:    code,
		Message: err.Error(),
	})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, autherrors.ErrUnauthenticated),
		errors.Is(err, autherrors.ErrInvalidToken),
		errors.Is(err, autherrors.ErrExpiredToken),
		errors.Is(err, autherrors.ErrBlacklistedToken),
		errors.Is(err, autherrors.ErrNoClaims),
		errors.Is(err, autherrors.ErrInvalidClaims):
		return nethttp.StatusUnauthorized, errorCode(err)
	case errors.Is(err, autherrors.ErrForbidden),
		errors.Is(err, autherrors.ErrForbiddenNotOwner),
		errors.Is(err, autherrors.ErrForbiddenWrongRole):
		return nethttp.StatusForbidden, errorCode(err)
	}
	return nethttp.StatusUnauthorized, "unauthenticated"
}

func errorCode(err error) string {
	switch {
	case errors.Is(err, autherrors.ErrExpiredToken):
		return "token_expired"
	case errors.Is(err, autherrors.ErrBlacklistedToken):
		return "token_revoked"
	case errors.Is(err, autherrors.ErrInvalidToken):
		return "invalid_token"
	case errors.Is(err, autherrors.ErrInvalidClaims):
		return "invalid_claims"
	case errors.Is(err, autherrors.ErrNoClaims):
		return "no_claims"
	case errors.Is(err, autherrors.ErrForbiddenNotOwner):
		return "forbidden_not_owner"
	case errors.Is(err, autherrors.ErrForbiddenWrongRole):
		return "forbidden_wrong_role"
	case errors.Is(err, autherrors.ErrForbidden):
		return "forbidden"
	}
	return "unauthenticated"
}
