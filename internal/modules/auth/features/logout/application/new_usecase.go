package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type LogoutUseCase struct {
	refreshTokenRepository RefreshTokenRepository
	blacklistRepository    BlacklistRepository
	sessionRepository      SessionRepository
	auditRepository        AuditRepository
	logger                 observability.Logger
}

type LogoutUseCaseDependencies struct {
	RefreshTokenRepository RefreshTokenRepository
	BlacklistRepository    BlacklistRepository
	SessionRepository      SessionRepository
	AuditRepository        AuditRepository
	Logger                 observability.Logger
}

func NewLogoutUseCase(deps LogoutUseCaseDependencies) *LogoutUseCase {
	return &LogoutUseCase{
		refreshTokenRepository: deps.RefreshTokenRepository,
		blacklistRepository:    deps.BlacklistRepository,
		sessionRepository:      deps.SessionRepository,
		auditRepository:        deps.AuditRepository,
		logger:                 deps.Logger,
	}
}
