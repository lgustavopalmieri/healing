package token

import (
	"context"
	"crypto/rsa"
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type JWTValidatorConfig struct {
	PublicKeys map[string]*rsa.PublicKey
	Issuer     string
	Audience   string
}

type JWTValidator struct {
	publicKeys map[string]*rsa.PublicKey
	issuer     string
	audience   string
}

func NewJWTValidator(cfg JWTValidatorConfig) *JWTValidator {
	return &JWTValidator{
		publicKeys: cfg.PublicKeys,
		issuer:     cfg.Issuer,
		audience:   cfg.Audience,
	}
}

type standardClaims struct {
	Role     string `json:"role"`
	Email    string `json:"email"`
	Provider string `json:"provider"`
	jwtlib.RegisteredClaims
}

func (v *JWTValidator) Validate(ctx context.Context, rawToken string) (*claims.Claims, error) {
	parsed, err := jwtlib.ParseWithClaims(rawToken, &standardClaims{}, rsaKeyFunc(v.publicKeys))
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, autherrors.ErrExpiredToken
		}
		return nil, autherrors.ErrInvalidToken
	}
	if !parsed.Valid {
		return nil, autherrors.ErrInvalidToken
	}
	sc, ok := parsed.Claims.(*standardClaims)
	if !ok {
		return nil, autherrors.ErrInvalidClaims
	}
	if sc.Issuer != v.issuer {
		return nil, autherrors.ErrInvalidClaims
	}
	if !containsAudience(sc.Audience, v.audience) {
		return nil, autherrors.ErrInvalidClaims
	}
	r, err := role.Parse(sc.Role)
	if err != nil {
		return nil, autherrors.ErrInvalidClaims
	}
	p, err := provider.Parse(sc.Provider)
	if err != nil {
		return nil, autherrors.ErrInvalidClaims
	}
	issuedAt := time.Time{}
	if sc.IssuedAt != nil {
		issuedAt = sc.IssuedAt.Time
	}
	expireAt := time.Time{}
	if sc.ExpiresAt != nil {
		expireAt = sc.ExpiresAt.Time
	}
	c := &claims.Claims{
		Subject:  sc.Subject,
		Role:     r,
		Email:    sc.Email,
		Provider: p,
		TokenID:  sc.ID,
		IssuedAt: issuedAt,
		ExpireAt: expireAt,
		Issuer:   sc.Issuer,
		Audience: firstAudience(sc.Audience),
	}
	if !c.Valid() {
		return nil, autherrors.ErrInvalidClaims
	}
	if c.IsExpired(time.Now()) {
		return nil, autherrors.ErrExpiredToken
	}
	return c, nil
}
