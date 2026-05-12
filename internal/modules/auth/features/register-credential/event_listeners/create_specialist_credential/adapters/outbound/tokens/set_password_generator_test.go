package tokens_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redistest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/cache/redis"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/create_specialist_credential/adapters/outbound/tokens"
	platformredis "github.com/lgustavopalmieri/healing-specialist/internal/platform/redis"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

const (
	testKID      = "healing-test-kid"
	testIssuer   = "healing-specialist"
	testAudience = "healing-platform"
	testTTL      = 24 * time.Hour
	testSubject  = "specialist-xyz-9"

	setPasswordKeyPrefix = "auth:set-password:"
)

type generatorTestContext struct {
	signer      *tokenissuer.Signer
	keyring     *tokenissuer.Keyring
	redisClient *goredis.Client
	redisAlive  bool
}

func newSigner(t *testing.T) (*tokenissuer.Signer, *tokenissuer.Keyring) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyring := tokenissuer.NewKeyring(testKID, privateKey)
	signer := tokenissuer.NewSigner(keyring, tokenissuer.SignerConfig{
		Issuer:   testIssuer,
		Audience: testAudience,
	})
	return signer, keyring
}

func setupGenerator(t *testing.T, killRedis bool) *generatorTestContext {
	t.Helper()
	signer, keyring := newSigner(t)

	container := redistest.SetupRedisContainer(t)
	client, err := platformredis.NewClient(context.Background(), container.Config())
	require.NoError(t, err)

	if killRedis {
		container.Terminate(t)
		_ = client.Close()
		return &generatorTestContext{
			signer:      signer,
			keyring:     keyring,
			redisClient: client,
			redisAlive:  false,
		}
	}

	t.Cleanup(func() { container.Terminate(t) })
	t.Cleanup(func() { _ = client.Close() })

	return &generatorTestContext{
		signer:      signer,
		keyring:     keyring,
		redisClient: client,
		redisAlive:  true,
	}
}

func TestSetPasswordTokenGenerator_Generate(t *testing.T) {
	tests := []struct {
		name           string
		killRedis      bool
		subject        string
		expectError    bool
		errMsgContains string
		validateResult func(t *testing.T, ctx *generatorTestContext, tokenString, jti string)
	}{
		{
			name:    "happy path - retorna JWT RS256 valido com purpose=set-password e jti nao vazio",
			subject: testSubject,
			validateResult: func(t *testing.T, gen *generatorTestContext, tokenString, jti string) {
				require.NotEmpty(t, tokenString)
				require.NotEmpty(t, jti)

				parsed, err := jwtlib.Parse(tokenString, func(token *jwtlib.Token) (interface{}, error) {
					kid, _ := token.Header["kid"].(string)
					return gen.keyring.PublicKeys[kid], nil
				})
				require.NoError(t, err)
				require.True(t, parsed.Valid)

				claims, ok := parsed.Claims.(jwtlib.MapClaims)
				require.True(t, ok)

				assert.Equal(t, jwtlib.SigningMethodRS256.Alg(), parsed.Method.Alg())
				assert.Equal(t, testKID, parsed.Header["kid"])
				assert.Equal(t, testSubject, claims["sub"])
				assert.Equal(t, "set-password", claims["purpose"])
				assert.Equal(t, jti, claims["jti"])
				assert.Equal(t, testIssuer, claims["iss"])
			},
		},
		{
			name:    "happy path - registra auth:set-password:{jti} no Redis com subjectID como valor",
			subject: testSubject,
			validateResult: func(t *testing.T, gen *generatorTestContext, _, jti string) {
				value, err := gen.redisClient.Get(context.Background(), setPasswordKeyPrefix+jti).Result()
				require.NoError(t, err)
				assert.Equal(t, testSubject, value)
			},
		},
		{
			name:    "happy path - registra chave com TTL proximo do configurado (margem de 1s)",
			subject: testSubject,
			validateResult: func(t *testing.T, gen *generatorTestContext, _, jti string) {
				ttl, err := gen.redisClient.TTL(context.Background(), setPasswordKeyPrefix+jti).Result()
				require.NoError(t, err)
				assert.InDelta(t, testTTL.Seconds(), ttl.Seconds(), 1.0)
			},
		},
		{
			name:           "failure - redis down retorna erro envelopando 'register set-password token'",
			subject:        testSubject,
			killRedis:      true,
			expectError:    true,
			errMsgContains: "register set-password token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := setupGenerator(t, tt.killRedis)
			generator := tokens.NewSetPasswordTokenGenerator(gen.signer, gen.redisClient, testTTL)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			tokenString, jti, err := generator.Generate(ctx, tt.subject)

			if tt.expectError {
				require.Error(t, err)
				if tt.errMsgContains != "" {
					assert.Contains(t, err.Error(), tt.errMsgContains)
				}
				assert.Empty(t, tokenString)
				assert.Empty(t, jti)
				return
			}
			require.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, gen, tokenString, jti)
			}
		})
	}
}
