package application

import (
	"context"
	"errors"
	"sync"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/audit"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/password"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/authutil"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/events"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	sdktoken "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
)

func (uc *SetPasswordUseCase) Execute(ctx context.Context, input SetPasswordDTO) (*SetPasswordResult, error) {
	validated, err := uc.tokenValidator.Validate(ctx, input.Token)
	if err != nil {
		return nil, ErrInvalidSetPasswordToken
	}

	consumed, err := uc.singleUseTokenRepository.Consume(ctx, validated.JTI)
	if err != nil {
		authutil.LogError(ctx, uc.logger, FailedToConsumeSingleUseTokenMessage, err, validated.SubjectID)
		return nil, ErrFailedToConsumeSingleUse
	}
	if !consumed {
		return nil, ErrSingleUseTokenAlreadyUsed
	}

	cred, err := uc.credentialRepository.FindBySubjectAndRole(ctx, validated.SubjectID, validated.Role)
	if err != nil {
		authutil.LogError(ctx, uc.logger, FailedToFindCredentialMessage, err, validated.SubjectID)
		return nil, ErrFailedToFindCredential
	}
	if cred == nil {
		return nil, ErrCredentialNotFound
	}
	if cred.Status != credential.StatusPending {
		return nil, ErrCredentialNotPending
	}

	pwd, err := password.NewPassword(input.Password, password.ValidationConfig{MinLength: uc.passwordMinLength})
	if err != nil {
		return nil, err
	}

	hashed, err := pwd.Hash(uc.bcryptCost)
	if err != nil {
		authutil.LogError(ctx, uc.logger, FailedToHashPasswordMessage, err, validated.SubjectID)
		return nil, ErrFailedToHashPassword
	}

	if err := cred.Activate(hashed); err != nil {
		return nil, errors.Join(ErrFailedToActivateCredential, err)
	}

	issued, err := uc.accessTokenIssuer.IssueAccessAndRefresh(ctx, cred)
	if err != nil {
		authutil.LogError(ctx, uc.logger, FailedToIssueTokenPairMessage, err, validated.SubjectID)
		return nil, ErrFailedToIssueTokenPair
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

	if err := uc.credentialRepository.UpdateWithSessionInTransaction(ctx, cred, sess); err != nil {
		authutil.LogError(ctx, uc.logger, FailedToPersistCredentialMessage, err, validated.SubjectID)
		return nil, ErrFailedToPersistCredential
	}

	if err := uc.refreshTokenRepository.Save(ctx, refreshTokenHash, refreshtoken.RefreshTokenPayload{
		SessionID: sess.ID,
		SubjectID: cred.SubjectID,
		Role:      cred.Role.String(),
		TTL:       authutil.RemainingTTL(issued.RefreshExpiresAt),
	}); err != nil {
		authutil.LogError(ctx, uc.logger, FailedToCacheRefreshTokenMessage, err, validated.SubjectID)
		return nil, ErrFailedToCacheRefreshToken
	}

	uc.runPostCommitSideEffects(ctx, cred, input)

	return &SetPasswordResult{
		TokenPair: issued,
		SubjectID: cred.SubjectID,
		Role:      cred.Role,
	}, nil
}

func (uc *SetPasswordUseCase) runPostCommitSideEffects(ctx context.Context, cred *credential.Credential, input SetPasswordDTO) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		uc.recordAuditPasswordSet(ctx, cred, input)
	}()

	go func() {
		defer wg.Done()
		uc.publishCredentialActivatedEvent(ctx, cred)
	}()

	wg.Wait()
}

func (uc *SetPasswordUseCase) recordAuditPasswordSet(ctx context.Context, cred *credential.Credential, input SetPasswordDTO) {
	log := audit.NewAuditLog(audit.NewAuditLogInput{
		SubjectID: cred.SubjectID,
		Role:      cred.Role,
		EventType: audit.EventPasswordSet,
		IPAddress: input.IPAddress,
		UserAgent: input.UserAgent,
	})
	if err := uc.auditRepository.Save(ctx, log); err != nil {
		authutil.LogError(ctx, uc.logger, FailedToRecordAuditMessage, err, cred.SubjectID)
	}
}

func (uc *SetPasswordUseCase) publishCredentialActivatedEvent(ctx context.Context, cred *credential.Credential) {
	evt := event.NewEvent(events.AuthCredentialActivated, map[string]any{
		"subject_id": cred.SubjectID,
		"role":       cred.Role.String(),
		"email":      cred.Email,
	})
	if err := uc.eventPublisher.Dispatch(ctx, evt); err != nil {
		authutil.LogError(ctx, uc.logger, FailedToPublishActivatedEventMessage, err, cred.SubjectID)
	}
}
