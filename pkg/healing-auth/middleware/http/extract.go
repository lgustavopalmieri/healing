package http

import (
	"strings"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
)

func ExtractBearerToken(headerValue string) (string, error) {
	if headerValue == "" {
		return "", autherrors.ErrUnauthenticated
	}
	if !strings.HasPrefix(headerValue, BearerPrefix) {
		return "", autherrors.ErrInvalidToken
	}
	raw := strings.TrimSpace(strings.TrimPrefix(headerValue, BearerPrefix))
	if raw == "" {
		return "", autherrors.ErrInvalidToken
	}
	return raw, nil
}
