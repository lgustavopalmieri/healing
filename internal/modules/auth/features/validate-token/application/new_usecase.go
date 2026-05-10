package application

import (
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/shared"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
)

type ValidateTokenUseCase struct {
	tokenValidator      token.TokenValidator
	blacklistRepository BlacklistRepository
}

func NewValidateTokenUseCase(
	tokenValidator token.TokenValidator,
	blacklistRepository BlacklistRepository,
) *ValidateTokenUseCase {
	return &ValidateTokenUseCase{
		tokenValidator:      tokenValidator,
		blacklistRepository: blacklistRepository,
	}
}

var _ shared.ValidateTokenUseCase = (*ValidateTokenUseCase)(nil)
