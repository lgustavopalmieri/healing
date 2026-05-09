package token_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token/testutil"
)

const (
	testIssuer    = "healing-specialist"
	testAudience  = "healing-platform"
	testKID       = "healing-2026-05"
	testOtherKID  = "healing-other-kid"
	testSubjectID = "subject-123"
	testEmail     = "specialist@healing.com"
	testTokenID   = "jti-abc"
)

type tokenClaims struct {
	Role     string `json:"role"`
	Email    string `json:"email"`
	Provider string `json:"provider"`
	jwtlib.RegisteredClaims
}

func defaultClaimsFactory(overrides ...func(*tokenClaims)) *tokenClaims {
	now := time.Now()
	c := &tokenClaims{
		Role:     role.Specialist.String(),
		Email:    testEmail,
		Provider: provider.Password.String(),
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   testSubjectID,
			Issuer:    testIssuer,
			Audience:  jwtlib.ClaimStrings{testAudience},
			ID:        testTokenID,
			IssuedAt:  jwtlib.NewNumericDate(now.Add(-1 * time.Minute)),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(1 * time.Hour)),
		},
	}
	for _, o := range overrides {
		o(c)
	}
	return c
}

func newValidator(t *testing.T, publicKey *rsa.PublicKey) *token.JWTValidator {
	t.Helper()
	return token.NewJWTValidator(token.JWTValidatorConfig{
		PublicKeys: map[string]*rsa.PublicKey{
			testKID: publicKey,
		},
		Issuer:   testIssuer,
		Audience: testAudience,
	})
}

func TestJWTValidator_Validate(t *testing.T) {
	privateKey, publicKey := testutil.GenerateRSAKeyPair(t)
	otherPrivateKey, _ := testutil.GenerateRSAKeyPair(t)

	tests := []struct {
		name        string
		tokenFn     func(t *testing.T) string
		expectError bool
		expectedErr error
		validate    func(t *testing.T, c *struct {
			Subject  string
			RoleStr  string
			Email    string
			Provider string
			TokenID  string
			Issuer   string
			Audience string
		})
	}{
		{
			name: "happy path - token valido com kid conhecido retorna claims corretas",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID:    testKID,
					Claims: defaultClaimsFactory(),
				})
			},
			expectError: false,
		},
		{
			name: "failure - kid desconhecido retorna ErrInvalidToken",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID:    testOtherKID,
					Claims: defaultClaimsFactory(),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
		{
			name: "failure - token sem kid no header retorna ErrInvalidToken",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					Claims: defaultClaimsFactory(),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
		{
			name: "failure - assinatura invalida (assinado com outra chave privada) retorna ErrInvalidToken",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, otherPrivateKey, testutil.SignOptions{
					KID:    testKID,
					Claims: defaultClaimsFactory(),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
		{
			name: "failure - metodo de assinatura HS256 retorna ErrInvalidToken",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenHS256(t, []byte("shared-secret"), testKID, defaultClaimsFactory())
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
		{
			name: "failure - token malformado retorna ErrInvalidToken",
			tokenFn: func(t *testing.T) string {
				return "not.a.valid.jwt"
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
		{
			name: "failure - issuer diferente do configurado retorna ErrInvalidClaims",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID: testKID,
					Claims: defaultClaimsFactory(func(c *tokenClaims) {
						c.Issuer = "other-issuer"
					}),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidClaims,
		},
		{
			name: "failure - audience diferente do configurado retorna ErrInvalidClaims",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID: testKID,
					Claims: defaultClaimsFactory(func(c *tokenClaims) {
						c.Audience = jwtlib.ClaimStrings{"other-audience"}
					}),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidClaims,
		},
		{
			name: "failure - audience ausente retorna ErrInvalidClaims",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID: testKID,
					Claims: defaultClaimsFactory(func(c *tokenClaims) {
						c.Audience = nil
					}),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidClaims,
		},
		{
			name: "failure - role invalida retorna ErrInvalidClaims",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID: testKID,
					Claims: defaultClaimsFactory(func(c *tokenClaims) {
						c.Role = "superadmin"
					}),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidClaims,
		},
		{
			name: "failure - role vazia retorna ErrInvalidClaims",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID: testKID,
					Claims: defaultClaimsFactory(func(c *tokenClaims) {
						c.Role = ""
					}),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidClaims,
		},
		{
			name: "failure - provider invalido retorna ErrInvalidClaims",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID: testKID,
					Claims: defaultClaimsFactory(func(c *tokenClaims) {
						c.Provider = "facebook"
					}),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidClaims,
		},
		{
			name: "failure - provider vazio retorna ErrInvalidClaims",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID: testKID,
					Claims: defaultClaimsFactory(func(c *tokenClaims) {
						c.Provider = ""
					}),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidClaims,
		},
		{
			name: "failure - subject vazio retorna ErrInvalidClaims",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID: testKID,
					Claims: defaultClaimsFactory(func(c *tokenClaims) {
						c.Subject = ""
					}),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidClaims,
		},
		{
			name: "failure - jti (TokenID) vazio retorna ErrInvalidClaims",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID: testKID,
					Claims: defaultClaimsFactory(func(c *tokenClaims) {
						c.ID = ""
					}),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidClaims,
		},
		{
			name: "failure - token expirado retorna ErrExpiredToken",
			tokenFn: func(t *testing.T) string {
				return testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
					KID: testKID,
					Claims: defaultClaimsFactory(func(c *tokenClaims) {
						now := time.Now()
						c.IssuedAt = jwtlib.NewNumericDate(now.Add(-2 * time.Hour))
						c.ExpiresAt = jwtlib.NewNumericDate(now.Add(-1 * time.Hour))
					}),
				})
			},
			expectError: true,
			expectedErr: autherrors.ErrExpiredToken,
		},
	}

	validator := newValidator(t, publicKey)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawToken := tt.tokenFn(t)

			claimsResult, err := validator.Validate(ctx, rawToken)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, claimsResult)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, claimsResult)
			assert.Equal(t, testSubjectID, claimsResult.Subject)
			assert.Equal(t, role.Specialist, claimsResult.Role)
			assert.Equal(t, testEmail, claimsResult.Email)
			assert.Equal(t, provider.Password, claimsResult.Provider)
			assert.Equal(t, testTokenID, claimsResult.TokenID)
			assert.Equal(t, testIssuer, claimsResult.Issuer)
			assert.Equal(t, testAudience, claimsResult.Audience)
			assert.False(t, claimsResult.IssuedAt.IsZero())
			assert.False(t, claimsResult.ExpireAt.IsZero())
		})
	}
}

func TestJWTValidator_Validate_MultipleAudiences(t *testing.T) {
	privateKey, publicKey := testutil.GenerateRSAKeyPair(t)
	validator := newValidator(t, publicKey)

	rawToken := testutil.SignTokenWithKID(t, privateKey, testutil.SignOptions{
		KID: testKID,
		Claims: defaultClaimsFactory(func(c *tokenClaims) {
			c.Audience = jwtlib.ClaimStrings{"first-audience", testAudience}
		}),
	})

	c, err := validator.Validate(context.Background(), rawToken)

	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "first-audience", c.Audience)
}

func TestNewJWTValidator(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	v := token.NewJWTValidator(token.JWTValidatorConfig{
		PublicKeys: map[string]*rsa.PublicKey{
			testKID: &key.PublicKey,
		},
		Issuer:   testIssuer,
		Audience: testAudience,
	})

	assert.NotNil(t, v)
}
