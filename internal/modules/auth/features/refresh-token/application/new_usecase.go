package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type RefreshTokenUseCase struct {
	refreshTokenRepository RefreshTokenRepository
	accessTokenIssuer      AccessTokenIssuer
	credentialRepository   CredentialRepository
	sessionRepository      SessionRepository
	logger                 observability.Logger
}

type RefreshTokenUseCaseDependencies struct {
	RefreshTokenRepository RefreshTokenRepository
	AccessTokenIssuer      AccessTokenIssuer
	CredentialRepository   CredentialRepository
	SessionRepository      SessionRepository
	Logger                 observability.Logger
}

func NewRefreshTokenUseCase(deps RefreshTokenUseCaseDependencies) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		refreshTokenRepository: deps.RefreshTokenRepository,
		accessTokenIssuer:      deps.AccessTokenIssuer,
		credentialRepository:   deps.CredentialRepository,
		sessionRepository:      deps.SessionRepository,
		logger:                 deps.Logger,
	}
}
