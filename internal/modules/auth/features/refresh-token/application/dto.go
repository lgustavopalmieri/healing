package application

import (
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type RefreshTokenDTO struct {
	RefreshToken string
}

type RefreshTokenResult struct {
	TokenPair *tokenpair.TokenPair
	SubjectID string
	Role      role.Role
}
