package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/audit"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

//go:generate mockgen -source=interface.go -destination=mocks/mocks.go -package=mocks
type CredentialRepository interface {
	FindByEmailProviderRole(ctx context.Context, email string, p provider.Provider, r role.Role) (*credential.Credential, error)
	UpdateLastUsed(ctx context.Context, credentialID string) error
}

type AccessTokenIssuer interface {
	IssueAccessAndRefresh(ctx context.Context, c *credential.Credential) (*tokenpair.TokenPair, error)
}

type SessionRepository interface {
	Save(ctx context.Context, s *session.Session) error
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, refreshTokenHash string, payload refreshtoken.RefreshTokenPayload) error
}

type LoginAttemptsTracker interface {
	Increment(ctx context.Context, email string) error
	Reset(ctx context.Context, email string) error
}

type AuditRepository interface {
	Save(ctx context.Context, log *audit.AuditLog) error
}
