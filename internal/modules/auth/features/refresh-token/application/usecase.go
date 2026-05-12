package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/authutil"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
	sdktoken "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
)

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshTokenDTO) (*RefreshTokenResult, error) {
	oldHash := sdktoken.Hash(input.RefreshToken)

	payload, err := uc.refreshTokenRepository.Find(ctx, oldHash)
	if err != nil {
		authutil.LogError(ctx, uc.logger, "failed to find refresh token", err, "")
		return nil, ErrInvalidRefreshToken
	}
	if payload == nil {
		return nil, ErrInvalidRefreshToken
	}

	if err := uc.refreshTokenRepository.Delete(ctx, oldHash); err != nil {
		authutil.LogError(ctx, uc.logger, "failed to delete old refresh token", err, payload.SubjectID)
		return nil, ErrDeleteOldRefresh
	}

	r, err := role.Parse(payload.Role)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	cred, err := uc.credentialRepository.FindBySubjectAndRole(ctx, payload.SubjectID, r)
	if err != nil || cred == nil {
		return nil, ErrInvalidRefreshToken
	}

	issued, err := uc.accessTokenIssuer.IssueAccessAndRefresh(ctx, cred)
	if err != nil {
		authutil.LogError(ctx, uc.logger, "failed to issue new token pair", err, payload.SubjectID)
		return nil, ErrIssueNewTokens
	}

	newHash := sdktoken.Hash(issued.RefreshToken)

	if err := uc.refreshTokenRepository.Save(ctx, newHash, refreshtoken.RefreshTokenPayload{
		SessionID: payload.SessionID,
		SubjectID: payload.SubjectID,
		Role:      payload.Role,
		TTL:       authutil.RemainingTTL(issued.RefreshExpiresAt),
	}); err != nil {
		authutil.LogError(ctx, uc.logger, "failed to cache new refresh token", err, payload.SubjectID)
		return nil, ErrCacheNewRefresh
	}

	_ = uc.sessionRepository.UpdateRefreshTokenHash(ctx, payload.SessionID, newHash)

	return &RefreshTokenResult{
		TokenPair: issued,
		SubjectID: payload.SubjectID,
		Role:      r,
	}, nil
}
