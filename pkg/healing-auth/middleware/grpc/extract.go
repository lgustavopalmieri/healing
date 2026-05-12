package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
)

const (
	AuthorizationMetadataKey = "authorization"
	BearerPrefix             = "Bearer "
)

func ExtractBearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", autherrors.ErrUnauthenticated
	}
	values := md.Get(AuthorizationMetadataKey)
	if len(values) == 0 {
		return "", autherrors.ErrUnauthenticated
	}
	header := values[0]
	if !strings.HasPrefix(header, BearerPrefix) {
		return "", autherrors.ErrInvalidToken
	}
	raw := strings.TrimSpace(strings.TrimPrefix(header, BearerPrefix))
	if raw == "" {
		return "", autherrors.ErrInvalidToken
	}
	return raw, nil
}
