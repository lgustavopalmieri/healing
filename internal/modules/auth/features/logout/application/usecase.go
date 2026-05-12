package application

import (
	"context"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/audit"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/authutil"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
	sdktoken "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
)

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutDTO) error {
	refreshHash := sdktoken.Hash(input.RefreshToken)

	if err := uc.refreshTokenRepository.Delete(ctx, refreshHash); err != nil {
		authutil.LogError(ctx, uc.logger, "failed to delete refresh token on logout", err, input.SubjectID)
		return ErrDeleteRefreshToken
	}

	ttl := time.Until(input.AccessTokenExp)
	if err := uc.blacklistRepository.Blacklist(ctx, input.AccessTokenJTI, ttl); err != nil {
		authutil.LogError(ctx, uc.logger, "failed to blacklist access token on logout", err, input.SubjectID)
		return ErrBlacklistAccessToken
	}

	r, _ := role.Parse(input.Role)
	_ = uc.auditRepository.Save(ctx, audit.NewAuditLog(audit.NewAuditLogInput{
		SubjectID: input.SubjectID,
		Role:      r,
		EventType: audit.EventLogout,
		IPAddress: input.IPAddress,
		UserAgent: input.UserAgent,
	}))

	return nil
}
