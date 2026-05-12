package token

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
)

type TokenValidator interface {
	Validate(ctx context.Context, rawToken string) (*claims.Claims, error)
}
