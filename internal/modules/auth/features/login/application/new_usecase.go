package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type LoginUseCase struct {
	credentialRepository   CredentialRepository
	accessTokenIssuer      AccessTokenIssuer
	sessionRepository      SessionRepository
	refreshTokenRepository RefreshTokenRepository
	loginAttemptsTracker   LoginAttemptsTracker
	auditRepository        AuditRepository
	logger                 observability.Logger
}

type LoginUseCaseDependencies struct {
	CredentialRepository   CredentialRepository
	AccessTokenIssuer      AccessTokenIssuer
	SessionRepository      SessionRepository
	RefreshTokenRepository RefreshTokenRepository
	LoginAttemptsTracker   LoginAttemptsTracker
	AuditRepository        AuditRepository
	Logger                 observability.Logger
}

func NewLoginUseCase(deps LoginUseCaseDependencies) *LoginUseCase {
	return &LoginUseCase{
		credentialRepository:   deps.CredentialRepository,
		accessTokenIssuer:      deps.AccessTokenIssuer,
		sessionRepository:      deps.SessionRepository,
		refreshTokenRepository: deps.RefreshTokenRepository,
		loginAttemptsTracker:   deps.LoginAttemptsTracker,
		auditRepository:        deps.AuditRepository,
		logger:                 deps.Logger,
	}
}
