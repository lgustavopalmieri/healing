package application

import (
	"context"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/audit"
)

//go:generate mockgen -source=interface.go -destination=mocks/mocks.go -package=mocks
type RefreshTokenRepository interface {
	Delete(ctx context.Context, refreshTokenHash string) error
}

type BlacklistRepository interface {
	Blacklist(ctx context.Context, jti string, ttl time.Duration) error
}

type SessionRepository interface {
	Revoke(ctx context.Context, sessionID string) error
}

type AuditRepository interface {
	Save(ctx context.Context, log *audit.AuditLog) error
}
