package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/audit"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

//go:generate mockgen -source=interface.go -destination=mocks/mocks.go -package=mocks
type SetPasswordTokenValidator interface {
	Validate(ctx context.Context, rawToken string) (*ValidatedSetPasswordToken, error)
}

type SingleUseTokenRepository interface {
	Consume(ctx context.Context, jti string) (bool, error)
}

type CredentialRepository interface {
	FindBySubjectAndRole(ctx context.Context, subjectID string, r role.Role) (*credential.Credential, error)
	UpdateWithSessionInTransaction(ctx context.Context, cred *credential.Credential, sess *session.Session) error
}

type AccessTokenIssuer interface {
	IssueAccessAndRefresh(ctx context.Context, c *credential.Credential) (*tokenpair.TokenPair, error)
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, refreshTokenHash string, payload refreshtoken.RefreshTokenPayload) error
}

type AuditRepository interface {
	Save(ctx context.Context, log *audit.AuditLog) error
}
