package tokenissuer

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type SignerConfig struct {
	Issuer   string
	Audience string
}

type Signer struct {
	keyring  *Keyring
	issuer   string
	audience string
}

func NewSigner(keyring *Keyring, cfg SignerConfig) *Signer {
	return &Signer{
		keyring:  keyring,
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
	}
}

type SignAccessInput struct {
	Subject  string
	Role     role.Role
	Email    string
	Provider provider.Provider
	TTL      time.Duration
}

type SignSpecialInput struct {
	Subject string
	Role    role.Role
	Purpose string
	TTL     time.Duration
}

type accessClaims struct {
	Role     string `json:"role"`
	Email    string `json:"email"`
	Provider string `json:"provider"`
	jwtlib.RegisteredClaims
}

type specialClaims struct {
	Role    string `json:"role"`
	Purpose string `json:"purpose"`
	jwtlib.RegisteredClaims
}

func (s *Signer) SignAccess(in SignAccessInput) (token string, jti string, exp time.Time, err error) {
	jti = uuid.New().String()
	now := time.Now()
	exp = now.Add(in.TTL)

	claims := accessClaims{
		Role:     in.Role.String(),
		Email:    in.Email,
		Provider: in.Provider.String(),
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   in.Subject,
			Issuer:    s.issuer,
			Audience:  jwtlib.ClaimStrings{s.audience},
			ID:        jti,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(exp),
		},
	}

	token, err = s.sign(claims)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("sign access: %w", err)
	}
	return token, jti, exp, nil
}

func (s *Signer) SignSpecialPurpose(in SignSpecialInput) (token string, jti string, exp time.Time, err error) {
	jti = uuid.New().String()
	now := time.Now()
	exp = now.Add(in.TTL)

	claims := specialClaims{
		Role:    in.Role.String(),
		Purpose: in.Purpose,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   in.Subject,
			Issuer:    s.issuer,
			Audience:  jwtlib.ClaimStrings{s.audience},
			ID:        jti,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(exp),
		},
	}

	token, err = s.sign(claims)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("sign special-purpose: %w", err)
	}
	return token, jti, exp, nil
}

func (s *Signer) sign(claims jwtlib.Claims) (string, error) {
	key, ok := s.keyring.PrivateKeys[s.keyring.CurrentKID]
	if !ok {
		return "", fmt.Errorf("current key %q not found in keyring", s.keyring.CurrentKID)
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyring.CurrentKID

	return token.SignedString(key)
}
