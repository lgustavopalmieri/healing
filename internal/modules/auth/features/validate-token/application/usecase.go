package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
)

func (uc *ValidateTokenUseCase) Execute(ctx context.Context, rawToken string) (*claims.Claims, error) {
	c, err := uc.tokenValidator.Validate(ctx, rawToken)
	if err != nil {
		return nil, err
	}

	if c.TokenID == "" {
		return c, nil
	}

	blacklisted, err := uc.blacklistRepository.IsBlacklisted(ctx, c.TokenID)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, autherrors.ErrBlacklistedToken
	}
	return c, nil
}
