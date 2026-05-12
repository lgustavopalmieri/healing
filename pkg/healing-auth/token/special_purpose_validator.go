package token

import (
	"context"
	"crypto/rsa"
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type SpecialPurposeJWTValidatorConfig struct {
	PublicKeys map[string]*rsa.PublicKey
	Issuer     string
	Audience   string
}

type SpecialPurposeJWTValidator struct {
	publicKeys map[string]*rsa.PublicKey
	issuer     string
	audience   string
}

func NewSpecialPurposeJWTValidator(cfg SpecialPurposeJWTValidatorConfig) *SpecialPurposeJWTValidator {
	return &SpecialPurposeJWTValidator{
		publicKeys: cfg.PublicKeys,
		issuer:     cfg.Issuer,
		audience:   cfg.Audience,
	}
}

type SpecialPurposeClaims struct {
	Subject  string
	Role     role.Role
	Purpose  string
	TokenID  string
	IssuedAt time.Time
	ExpireAt time.Time
	Issuer   string
	Audience string
}

type specialPurposeJWTClaims struct {
	Role    string `json:"role"`
	Purpose string `json:"purpose"`
	jwtlib.RegisteredClaims
}

func (v *SpecialPurposeJWTValidator) Validate(ctx context.Context, rawToken string, expectedPurpose string) (*SpecialPurposeClaims, error) {
	parsed, err := jwtlib.ParseWithClaims(rawToken, &specialPurposeJWTClaims{}, rsaKeyFunc(v.publicKeys))
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, autherrors.ErrExpiredToken
		}
		return nil, autherrors.ErrInvalidToken
	}
	if !parsed.Valid {
		return nil, autherrors.ErrInvalidToken
	}
	sc, ok := parsed.Claims.(*specialPurposeJWTClaims)
	if !ok {
		return nil, autherrors.ErrInvalidClaims
	}
	if sc.Issuer != v.issuer {
		return nil, autherrors.ErrInvalidClaims
	}
	if !containsAudience(sc.Audience, v.audience) {
		return nil, autherrors.ErrInvalidClaims
	}
	if sc.Purpose == "" || sc.Purpose != expectedPurpose {
		return nil, autherrors.ErrInvalidClaims
	}
	r, err := role.Parse(sc.Role)
	if err != nil {
		return nil, autherrors.ErrInvalidClaims
	}
	if sc.Subject == "" || sc.ID == "" {
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
	if time.Now().After(expireAt) {
		return nil, autherrors.ErrExpiredToken
	}

	return &SpecialPurposeClaims{
		Subject:  sc.Subject,
		Role:     r,
		Purpose:  sc.Purpose,
		TokenID:  sc.ID,
		IssuedAt: issuedAt,
		ExpireAt: expireAt,
		Issuer:   sc.Issuer,
		Audience: firstAudience(sc.Audience),
	}, nil
}
