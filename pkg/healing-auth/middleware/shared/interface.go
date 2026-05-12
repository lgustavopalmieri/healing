package shared

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
)

//go:generate mockgen -source=interface.go -destination=mocks/validate_token_mock.go -package=mocks
type ValidateTokenUseCase interface {
	Execute(ctx context.Context, rawToken string) (*claims.Claims, error)
}
