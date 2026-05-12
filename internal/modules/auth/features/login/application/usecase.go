package application

import (
	"context"
	"sync"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/audit"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/password"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/authutil"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
	sdktoken "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
)

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginDTO) (*LoginResult, error) {
	expectedRole, err := role.Parse(input.ExpectedRole)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	cred, err := uc.credentialRepository.FindByEmailProviderRole(ctx, input.Email, provider.Password, expectedRole)
	if err != nil {
		authutil.LogError(ctx, uc.logger, FailedToFindCredentialMessage, err, "")
		uc.auditFailure(ctx, input, "not_found")
		return nil, ErrInvalidCredentials
	}
	if cred == nil {
		uc.auditFailure(ctx, input, "not_found")
		return nil, ErrInvalidCredentials
	}

	switch cred.Status {
	case credential.StatusLocked:
		uc.auditFailure(ctx, input, "locked")
		return nil, ErrCredentialLocked
	case credential.StatusActive:
		// ok
	default:
		uc.auditFailure(ctx, input, "status_"+string(cred.Status))
		return nil, ErrInvalidCredentials
	}

	attempt, err := password.NewPassword(input.Password, password.ValidationConfig{MinLength: 1})
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !attempt.Matches(cred.PasswordHash) {
		_ = uc.loginAttemptsTracker.Increment(ctx, input.Email)
		uc.auditFailure(ctx, input, "invalid_password")
		return nil, ErrInvalidCredentials
	}

	issued, err := uc.accessTokenIssuer.IssueAccessAndRefresh(ctx, cred)
	if err != nil {
		authutil.LogError(ctx, uc.logger, FailedToIssueTokenPairMessage, err, cred.SubjectID)
		return nil, ErrIssueTokens
	}

	refreshTokenHash := sdktoken.Hash(issued.RefreshToken)

	sess := session.NewSession(session.NewSessionInput{
		SubjectID:        cred.SubjectID,
		Role:             cred.Role,
		RefreshTokenHash: refreshTokenHash,
		DeviceInfo:       input.DeviceInfo,
		IPAddress:        input.IPAddress,
		UserAgent:        input.UserAgent,
		ExpiresAt:        issued.RefreshExpiresAt,
	})

	if err := uc.sessionRepository.Save(ctx, sess); err != nil {
		authutil.LogError(ctx, uc.logger, FailedToPersistSessionMessage, err, cred.SubjectID)
		return nil, ErrPersistSession
	}

	if err := uc.refreshTokenRepository.Save(ctx, refreshTokenHash, refreshtoken.RefreshTokenPayload{
		SessionID: sess.ID,
		SubjectID: cred.SubjectID,
		Role:      cred.Role.String(),
		TTL:       authutil.RemainingTTL(issued.RefreshExpiresAt),
	}); err != nil {
		authutil.LogError(ctx, uc.logger, FailedToCacheRefreshMessage, err, cred.SubjectID)
		return nil, ErrCacheRefreshToken
	}

	uc.runPostLoginSideEffects(ctx, cred, input)

	return &LoginResult{
		TokenPair: issued,
		SubjectID: cred.SubjectID,
		Role:      cred.Role,
	}, nil
}

func (uc *LoginUseCase) runPostLoginSideEffects(ctx context.Context, cred *credential.Credential, input LoginDTO) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		_ = uc.loginAttemptsTracker.Reset(ctx, input.Email)
	}()

	go func() {
		defer wg.Done()
		_ = uc.credentialRepository.UpdateLastUsed(ctx, cred.ID)
	}()

	go func() {
		defer wg.Done()
		_ = uc.auditRepository.Save(ctx, audit.NewAuditLog(audit.NewAuditLogInput{
			SubjectID: cred.SubjectID,
			Role:      cred.Role,
			EventType: audit.EventLoginSuccess,
			IPAddress: input.IPAddress,
			UserAgent: input.UserAgent,
		}))
	}()

	wg.Wait()
}

func (uc *LoginUseCase) auditFailure(ctx context.Context, input LoginDTO, reason string) {
	_ = uc.auditRepository.Save(ctx, audit.NewAuditLog(audit.NewAuditLogInput{
		EventType: audit.EventLoginFailure,
		IPAddress: input.IPAddress,
		UserAgent: input.UserAgent,
		Metadata:  map[string]any{"reason": reason, "email": input.Email},
	}))
}
