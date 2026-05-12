package httphandler

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/refresh-token/application"
	refreshtokenrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
)

type Dependencies struct {
	RefreshTokenRepository application.RefreshTokenRepository
	AccessTokenIssuer      application.AccessTokenIssuer
	CredentialRepository   application.CredentialRepository
	SessionRepository      application.SessionRepository
	Logger                 observability.Logger
}

func NewRefreshTokenHTTPHandlerFromDeps(deps Dependencies) *RefreshTokenHTTPHandler {
	uc := application.NewRefreshTokenUseCase(application.RefreshTokenUseCaseDependencies{
		RefreshTokenRepository: deps.RefreshTokenRepository,
		AccessTokenIssuer:      deps.AccessTokenIssuer,
		CredentialRepository:   deps.CredentialRepository,
		SessionRepository:      deps.SessionRepository,
		Logger:                 deps.Logger,
	})
	return NewRefreshTokenHTTPHandler(uc)
}

var _ application.RefreshTokenRepository = (*refreshtokenrepo.RefreshTokenCacheRepository)(nil)
