package tokens

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

const (
	setPasswordPurpose   = "set-password"
	setPasswordKeyPrefix = "auth:set-password:"
)

type SetPasswordTokenGenerator struct {
	signer      *tokenissuer.Signer
	redisClient *redis.Client
	tokenTTL    time.Duration
}

func NewSetPasswordTokenGenerator(
	signer *tokenissuer.Signer,
	redisClient *redis.Client,
	tokenTTL time.Duration,
) *SetPasswordTokenGenerator {
	return &SetPasswordTokenGenerator{
		signer:      signer,
		redisClient: redisClient,
		tokenTTL:    tokenTTL,
	}
}

func (g *SetPasswordTokenGenerator) Generate(ctx context.Context, subjectID string) (string, string, error) {
	tokenString, jti, _, err := g.signer.SignSpecialPurpose(tokenissuer.SignSpecialInput{
		Subject: subjectID,
		Role:    role.Specialist,
		Purpose: setPasswordPurpose,
		TTL:     g.tokenTTL,
	})
	if err != nil {
		return "", "", fmt.Errorf("sign set-password token: %w", err)
	}

	key := setPasswordKeyPrefix + jti
	if err := g.redisClient.Set(ctx, key, subjectID, g.tokenTTL).Err(); err != nil {
		return "", "", fmt.Errorf("register set-password token: %w", err)
	}

	return tokenString, jti, nil
}
