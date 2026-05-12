package provider_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/adapters/outbound/provider"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

const (
	testIssuer   = "healing-specialist"
	testAudience = "healing-platform"
	testKID      = "healing-test-kid"
)

func newValidator(t *testing.T) (*provider.SetPasswordTokenValidator, *tokenissuer.Keyring) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyring := tokenissuer.NewKeyring(testKID, key)
	validator := provider.NewSetPasswordTokenValidator(provider.SetPasswordTokenValidatorConfig{
		Keyring:  keyring,
		Issuer:   testIssuer,
		Audience: testAudience,
	})
	return validator, keyring
}

func signSpecialPurposeToken(t *testing.T, keyring *tokenissuer.Keyring, claims jwtlib.Claims) string {
	t.Helper()
	signer := tokenissuer.NewSigner(keyring, tokenissuer.SignerConfig{
		Issuer:   testIssuer,
		Audience: testAudience,
	})

	token, _, _, err := signer.SignSpecialPurpose(tokenissuer.SignSpecialInput{
		Subject: "subject-1",
		Role:    role.Specialist,
		Purpose: "set-password",
		TTL:     24 * time.Hour,
	})
	require.NoError(t, err)
	return token
}

type specialClaims struct {
	Role    string `json:"role"`
	Purpose string `json:"purpose"`
	jwtlib.RegisteredClaims
}

func customSignedToken(t *testing.T, kr *tokenissuer.Keyring, kid string, signKey *rsa.PrivateKey, claims specialClaims) string {
	t.Helper()
	method := jwtlib.SigningMethodRS256
	tok := jwtlib.NewWithClaims(method, claims)
	tok.Header["kid"] = kid

	key := signKey
	if key == nil {
		key = kr.PrivateKeys[kr.CurrentKID]
	}

	signed, err := tok.SignedString(key)
	require.NoError(t, err)
	return signed
}

func TestSetPasswordTokenValidator_Validate(t *testing.T) {
	validator, keyring := newValidator(t)
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tests := []struct {
		name        string
		tokenFn     func(t *testing.T) string
		expectError bool
		expectedErr error
	}{
		{
			name: "happy path - token RS256 com purpose=set-password retorna claims validas",
			tokenFn: func(t *testing.T) string {
				return signSpecialPurposeToken(t, keyring, nil)
			},
		},
		{
			name: "failure - purpose diferente retorna erro (ErrInvalidClaims)",
			tokenFn: func(t *testing.T) string {
				now := time.Now()
				return customSignedToken(t, keyring, testKID, nil, specialClaims{
					Role:    role.Specialist.String(),
					Purpose: "reset-password",
					RegisteredClaims: jwtlib.RegisteredClaims{
						Subject:   "subject-1",
						Issuer:    testIssuer,
						Audience:  jwtlib.ClaimStrings{testAudience},
						ID:        "jti-purpose-mismatch",
						IssuedAt:  jwtlib.NewNumericDate(now),
						ExpiresAt: jwtlib.NewNumericDate(now.Add(1 * time.Hour)),
					},
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidClaims,
		},
		{
			name: "failure - token expirado retorna ErrExpiredToken",
			tokenFn: func(t *testing.T) string {
				now := time.Now()
				return customSignedToken(t, keyring, testKID, nil, specialClaims{
					Role:    role.Specialist.String(),
					Purpose: "set-password",
					RegisteredClaims: jwtlib.RegisteredClaims{
						Subject:   "subject-1",
						Issuer:    testIssuer,
						Audience:  jwtlib.ClaimStrings{testAudience},
						ID:        "jti-expired",
						IssuedAt:  jwtlib.NewNumericDate(now.Add(-2 * time.Hour)),
						ExpiresAt: jwtlib.NewNumericDate(now.Add(-1 * time.Hour)),
					},
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrExpiredToken,
		},
		{
			name: "failure - kid desconhecido retorna ErrInvalidToken",
			tokenFn: func(t *testing.T) string {
				now := time.Now()
				return customSignedToken(t, keyring, "unknown-kid", nil, specialClaims{
					Role:    role.Specialist.String(),
					Purpose: "set-password",
					RegisteredClaims: jwtlib.RegisteredClaims{
						Subject:   "subject-1",
						Issuer:    testIssuer,
						Audience:  jwtlib.ClaimStrings{testAudience},
						ID:        "jti-unknown-kid",
						IssuedAt:  jwtlib.NewNumericDate(now),
						ExpiresAt: jwtlib.NewNumericDate(now.Add(1 * time.Hour)),
					},
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
		{
			name: "failure - assinado com chave diferente retorna ErrInvalidToken",
			tokenFn: func(t *testing.T) string {
				now := time.Now()
				return customSignedToken(t, keyring, testKID, otherKey, specialClaims{
					Role:    role.Specialist.String(),
					Purpose: "set-password",
					RegisteredClaims: jwtlib.RegisteredClaims{
						Subject:   "subject-1",
						Issuer:    testIssuer,
						Audience:  jwtlib.ClaimStrings{testAudience},
						ID:        "jti-wrong-key",
						IssuedAt:  jwtlib.NewNumericDate(now),
						ExpiresAt: jwtlib.NewNumericDate(now.Add(1 * time.Hour)),
					},
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validator.Validate(context.Background(), tt.tokenFn(t))

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, "subject-1", got.SubjectID)
			assert.Equal(t, role.Specialist, got.Role)
			assert.NotEmpty(t, got.JTI)
		})
	}
}
