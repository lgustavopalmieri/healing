package tokenissuer_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
	sdktoken "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
)

const (
	testIssuer    = "healing-specialist"
	testAudience  = "healing-platform"
	testKID       = "healing-2026-05"
	testSubjectID = "subject-123"
	testEmail     = "specialist@healing.com"
)

func newSigner(t *testing.T) (*tokenissuer.Signer, *tokenissuer.Keyring, *rsa.PrivateKey) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyring := tokenissuer.NewKeyring(testKID, key)
	signer := tokenissuer.NewSigner(keyring, tokenissuer.SignerConfig{
		Issuer:   testIssuer,
		Audience: testAudience,
	})
	return signer, keyring, key
}

func keyFuncFromKeyring(kr *tokenissuer.Keyring) jwtlib.Keyfunc {
	return func(t *jwtlib.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		return kr.PublicKeys[kid], nil
	}
}

func TestSigner_SignAccess(t *testing.T) {
	t.Run("happy path - gera token RS256 com kid no header e claims corretos", func(t *testing.T) {
		signer, keyring, _ := newSigner(t)

		start := time.Now()
		tokenStr, jti, exp, err := signer.SignAccess(tokenissuer.SignAccessInput{
			Subject:  testSubjectID,
			Role:     role.Specialist,
			Email:    testEmail,
			Provider: provider.Password,
			TTL:      1 * time.Hour,
		})

		require.NoError(t, err)
		require.NotEmpty(t, tokenStr)
		require.NotEmpty(t, jti)
		assert.WithinDuration(t, start.Add(1*time.Hour), exp, 2*time.Second)

		parsed, err := jwtlib.Parse(tokenStr, keyFuncFromKeyring(keyring))
		require.NoError(t, err)
		require.True(t, parsed.Valid)

		assert.Equal(t, jwtlib.SigningMethodRS256.Alg(), parsed.Method.Alg())
		assert.Equal(t, testKID, parsed.Header["kid"])

		claims, ok := parsed.Claims.(jwtlib.MapClaims)
		require.True(t, ok)

		assert.Equal(t, testSubjectID, claims["sub"])
		assert.Equal(t, role.Specialist.String(), claims["role"])
		assert.Equal(t, testEmail, claims["email"])
		assert.Equal(t, provider.Password.String(), claims["provider"])
		assert.Equal(t, jti, claims["jti"])
		assert.Equal(t, testIssuer, claims["iss"])

		aud, ok := claims["aud"].([]interface{})
		require.True(t, ok)
		require.Len(t, aud, 1)
		assert.Equal(t, testAudience, aud[0])

		iat, ok := claims["iat"].(float64)
		require.True(t, ok)
		expClaim, ok := claims["exp"].(float64)
		require.True(t, ok)
		assert.InDelta(t, exp.Unix(), int64(expClaim), 1)
		assert.InDelta(t, 3600, expClaim-iat, 2)
	})

	t.Run("failure - current kid ausente no keyring retorna erro", func(t *testing.T) {
		keyring := &tokenissuer.Keyring{
			CurrentKID:  "missing-kid",
			PrivateKeys: map[string]*rsa.PrivateKey{},
			PublicKeys:  map[string]*rsa.PublicKey{},
		}
		signer := tokenissuer.NewSigner(keyring, tokenissuer.SignerConfig{
			Issuer:   testIssuer,
			Audience: testAudience,
		})

		tokenStr, jti, exp, err := signer.SignAccess(tokenissuer.SignAccessInput{
			Subject:  testSubjectID,
			Role:     role.Specialist,
			Email:    testEmail,
			Provider: provider.Password,
			TTL:      1 * time.Hour,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "current key")
		assert.Contains(t, err.Error(), "not found")
		assert.Empty(t, tokenStr)
		assert.Empty(t, jti)
		assert.True(t, exp.IsZero())
	})
}

func TestSigner_SignSpecialPurpose(t *testing.T) {
	t.Run("happy path - gera token com claim 'purpose' preenchido", func(t *testing.T) {
		signer, keyring, _ := newSigner(t)

		start := time.Now()
		tokenStr, jti, exp, err := signer.SignSpecialPurpose(tokenissuer.SignSpecialInput{
			Subject: testSubjectID,
			Role:    role.Specialist,
			Purpose: "set-password",
			TTL:     24 * time.Hour,
		})

		require.NoError(t, err)
		require.NotEmpty(t, tokenStr)
		require.NotEmpty(t, jti)
		assert.WithinDuration(t, start.Add(24*time.Hour), exp, 2*time.Second)

		parsed, err := jwtlib.Parse(tokenStr, keyFuncFromKeyring(keyring))
		require.NoError(t, err)
		require.True(t, parsed.Valid)
		assert.Equal(t, testKID, parsed.Header["kid"])

		claims, ok := parsed.Claims.(jwtlib.MapClaims)
		require.True(t, ok)

		assert.Equal(t, testSubjectID, claims["sub"])
		assert.Equal(t, role.Specialist.String(), claims["role"])
		assert.Equal(t, "set-password", claims["purpose"])
		assert.Equal(t, jti, claims["jti"])
		assert.Equal(t, testIssuer, claims["iss"])
	})

	t.Run("failure - current kid ausente no keyring retorna erro", func(t *testing.T) {
		keyring := &tokenissuer.Keyring{
			CurrentKID:  "missing-kid",
			PrivateKeys: map[string]*rsa.PrivateKey{},
			PublicKeys:  map[string]*rsa.PublicKey{},
		}
		signer := tokenissuer.NewSigner(keyring, tokenissuer.SignerConfig{
			Issuer:   testIssuer,
			Audience: testAudience,
		})

		tokenStr, jti, exp, err := signer.SignSpecialPurpose(tokenissuer.SignSpecialInput{
			Subject: testSubjectID,
			Role:    role.Specialist,
			Purpose: "set-password",
			TTL:     1 * time.Hour,
		})

		require.Error(t, err)
		assert.Empty(t, tokenStr)
		assert.Empty(t, jti)
		assert.True(t, exp.IsZero())
	})
}

func TestSigner_RoundTripWithJWTValidator(t *testing.T) {
	t.Run("round-trip - token assinado pelo Signer e validado pelo SDK JWTValidator", func(t *testing.T) {
		signer, keyring, _ := newSigner(t)

		tokenStr, jti, _, err := signer.SignAccess(tokenissuer.SignAccessInput{
			Subject:  testSubjectID,
			Role:     role.Specialist,
			Email:    testEmail,
			Provider: provider.Password,
			TTL:      1 * time.Hour,
		})
		require.NoError(t, err)

		validator := sdktoken.NewJWTValidator(sdktoken.JWTValidatorConfig{
			PublicKeys: keyring.PublicKeys,
			Issuer:     testIssuer,
			Audience:   testAudience,
		})

		claims, err := validator.Validate(context.Background(), tokenStr)
		require.NoError(t, err)
		require.NotNil(t, claims)

		assert.Equal(t, testSubjectID, claims.Subject)
		assert.Equal(t, role.Specialist, claims.Role)
		assert.Equal(t, testEmail, claims.Email)
		assert.Equal(t, provider.Password, claims.Provider)
		assert.Equal(t, jti, claims.TokenID)
		assert.Equal(t, testIssuer, claims.Issuer)
		assert.Equal(t, testAudience, claims.Audience)
	})
}
