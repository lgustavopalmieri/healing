package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type SetPasswordUseCase struct {
	tokenValidator           SetPasswordTokenValidator
	singleUseTokenRepository SingleUseTokenRepository
	credentialRepository     CredentialRepository
	accessTokenIssuer        AccessTokenIssuer
	refreshTokenRepository   RefreshTokenRepository
	auditRepository          AuditRepository
	eventPublisher           event.EventDispatcher
	logger                   observability.Logger
	passwordMinLength        int
	bcryptCost               int
}

type SetPasswordUseCaseDependencies struct {
	TokenValidator           SetPasswordTokenValidator
	SingleUseTokenRepository SingleUseTokenRepository
	CredentialRepository     CredentialRepository
	AccessTokenIssuer        AccessTokenIssuer
	RefreshTokenRepository   RefreshTokenRepository
	AuditRepository          AuditRepository
	EventPublisher           event.EventDispatcher
	Logger                   observability.Logger
	PasswordMinLength        int
	BcryptCost               int
}

func NewSetPasswordUseCase(deps SetPasswordUseCaseDependencies) *SetPasswordUseCase {
	return &SetPasswordUseCase{
		tokenValidator:           deps.TokenValidator,
		singleUseTokenRepository: deps.SingleUseTokenRepository,
		credentialRepository:     deps.CredentialRepository,
		accessTokenIssuer:        deps.AccessTokenIssuer,
		refreshTokenRepository:   deps.RefreshTokenRepository,
		auditRepository:          deps.AuditRepository,
		eventPublisher:           deps.EventPublisher,
		logger:                   deps.Logger,
		passwordMinLength:        deps.PasswordMinLength,
		bcryptCost:               deps.BcryptCost,
	}
}
