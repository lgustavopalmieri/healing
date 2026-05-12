package accesstokenissuer_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	accesstokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/providers/access_token_issuer"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

const (
	testKID      = "healing-test-kid"
	testIssuer   = "healing-specialist"
	testAudience = "healing-platform"
)

func newIssuer(t *testing.T, accessTTL, refreshTTL time.Duration) (*accesstokenissuer.AccessTokenIssuer, *tokenissuer.Keyring) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyring := tokenissuer.NewKeyring(testKID, key)
	signer := tokenissuer.NewSigner(keyring, tokenissuer.SignerConfig{
		Issuer:   testIssuer,
		Audience: testAudience,
	})

	issuer := accesstokenissuer.NewAccessTokenIssuer(accesstokenissuer.AccessTokenIssuerConfig{
		Signer:          signer,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
	})
	return issuer, keyring
}

func credentialFactory() *credential.Credential {
	return &credential.Credential{
		ID:        "cred-1",
		SubjectID: "subject-1",
		Role:      role.Specialist,
		Provider:  provider.Password,
		Email:     "user@healing.com",
		Status:    credential.StatusActive,
	}
}

func TestAccessTokenIssuer_IssueAccessAndRefresh(t *testing.T) {
	tests := []struct {
		name       string
		accessTTL  time.Duration
		refreshTTL time.Duration
		validate   func(t *testing.T, keyring *tokenissuer.Keyring, accessTTL, refreshTTL time.Duration, pair any)
	}{
		{
			name:       "happy path - emite access JWT RS256 com kid, iss, aud, sub, role, email, provider, jti",
			accessTTL:  1 * time.Hour,
			refreshTTL: 168 * time.Hour,
		},
		{
			name:       "happy path - refresh token e string opaca base64 url-safe diferente do access",
			accessTTL:  30 * time.Minute,
			refreshTTL: 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issuer, keyring := newIssuer(t, tt.accessTTL, tt.refreshTTL)
			cred := credentialFactory()
			start := time.Now()

			pair, err := issuer.IssueAccessAndRefresh(context.Background(), cred)

			require.NoError(t, err)
			require.NotNil(t, pair)
			require.NotEmpty(t, pair.AccessToken)
			require.NotEmpty(t, pair.AccessJTI)
			require.NotEmpty(t, pair.RefreshToken)
			assert.NotEqual(t, pair.AccessToken, pair.RefreshToken)

			parsed, err := jwtlib.Parse(pair.AccessToken, func(tok *jwtlib.Token) (interface{}, error) {
				kid, _ := tok.Header["kid"].(string)
				return keyring.PublicKeys[kid], nil
			})
			require.NoError(t, err)
			require.True(t, parsed.Valid)
			assert.Equal(t, testKID, parsed.Header["kid"])

			claims := parsed.Claims.(jwtlib.MapClaims)
			assert.Equal(t, cred.SubjectID, claims["sub"])
			assert.Equal(t, role.Specialist.String(), claims["role"])
			assert.Equal(t, cred.Email, claims["email"])
			assert.Equal(t, provider.Password.String(), claims["provider"])
			assert.Equal(t, testIssuer, claims["iss"])
			assert.Equal(t, pair.AccessJTI, claims["jti"])

			assert.WithinDuration(t, start.Add(tt.accessTTL), pair.AccessExpiresAt, 2*time.Second)
			assert.WithinDuration(t, start.Add(tt.refreshTTL), pair.RefreshExpiresAt, 2*time.Second)

			decoded, err := base64.RawURLEncoding.DecodeString(pair.RefreshToken)
			require.NoError(t, err, "refresh token should be base64 url-safe")
			assert.Equal(t, 32, len(decoded), "refresh token should be 32 raw bytes")
		})
	}
}
