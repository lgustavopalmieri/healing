package claims

import (
	"context"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
)

type ctxKey struct{}

func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

func FromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*Claims)
	return c, ok && c != nil
}

func MustFromContext(ctx context.Context) (*Claims, error) {
	c, ok := FromContext(ctx)
	if !ok {
		return nil, autherrors.ErrNoClaims
	}
	if !c.Valid() {
		return nil, autherrors.ErrInvalidClaims
	}
	return c, nil
}
