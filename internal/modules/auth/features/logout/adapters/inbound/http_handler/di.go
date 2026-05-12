package httphandler

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/logout/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/blacklist"
)

type Dependencies struct {
	RefreshTokenRepository application.RefreshTokenRepository
	BlacklistRepository    application.BlacklistRepository
	SessionRepository      application.SessionRepository
	AuditRepository        application.AuditRepository
	Logger                 observability.Logger
}

func NewLogoutHTTPHandlerFromDeps(deps Dependencies) *LogoutHTTPHandler {
	uc := application.NewLogoutUseCase(application.LogoutUseCaseDependencies{
		RefreshTokenRepository: deps.RefreshTokenRepository,
		BlacklistRepository:    deps.BlacklistRepository,
		SessionRepository:      deps.SessionRepository,
		AuditRepository:        deps.AuditRepository,
		Logger:                 deps.Logger,
	})
	return NewLogoutHTTPHandler(uc)
}

var _ application.BlacklistRepository = (*blacklist.BlacklistCacheRepository)(nil)
