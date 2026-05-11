package accesstokenissuer

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

const refreshTokenBytes = 32

type AccessTokenIssuerConfig struct {
	Signer          *tokenissuer.Signer
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type AccessTokenIssuer struct {
	signer          *tokenissuer.Signer
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAccessTokenIssuer(cfg AccessTokenIssuerConfig) *AccessTokenIssuer {
	return &AccessTokenIssuer{
		signer:          cfg.Signer,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func (i *AccessTokenIssuer) IssueAccessAndRefresh(ctx context.Context, c *credential.Credential) (*tokenpair.TokenPair, error) {
	accessToken, accessJTI, accessExp, err := i.signer.SignAccess(tokenissuer.SignAccessInput{
		Subject:  c.SubjectID,
		Role:     c.Role,
		Email:    c.Email,
		Provider: c.Provider,
		TTL:      i.accessTokenTTL,
	})
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshToken, err := generateOpaqueToken(refreshTokenBytes)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &tokenpair.TokenPair{
		AccessToken:      accessToken,
		AccessJTI:        accessJTI,
		AccessExpiresAt:  accessExp,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: time.Now().Add(i.refreshTokenTTL),
	}, nil
}

func generateOpaqueToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
