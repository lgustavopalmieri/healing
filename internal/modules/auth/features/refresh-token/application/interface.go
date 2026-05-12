package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

//go:generate mockgen -source=interface.go -destination=mocks/mocks.go -package=mocks
type RefreshTokenRepository interface {
	Find(ctx context.Context, refreshTokenHash string) (*refreshtoken.RefreshTokenPayload, error)
	Delete(ctx context.Context, refreshTokenHash string) error
	Save(ctx context.Context, refreshTokenHash string, payload refreshtoken.RefreshTokenPayload) error
}

type AccessTokenIssuer interface {
	IssueAccessAndRefresh(ctx context.Context, c *credential.Credential) (*tokenpair.TokenPair, error)
}

type CredentialRepository interface {
	FindBySubjectAndRole(ctx context.Context, subjectID string, r role.Role) (*credential.Credential, error)
}

type SessionRepository interface {
	UpdateRefreshTokenHash(ctx context.Context, sessionID, newHash string) error
}
