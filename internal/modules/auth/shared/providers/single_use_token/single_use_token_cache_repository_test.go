package singleusetoken_test

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redistest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/cache/redis"
	singleusetoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/providers/single_use_token"
	platformredis "github.com/lgustavopalmieri/healing-specialist/internal/platform/redis"
)

func setupRepository(t *testing.T) (*singleusetoken.SingleUseTokenCacheRepository, *goredis.Client, *redistest.RedisContainer) {
	t.Helper()
	container := redistest.SetupRedisContainer(t)
	client, err := platformredis.NewClient(context.Background(), container.Config())
	require.NoError(t, err)
	return singleusetoken.NewSingleUseTokenCacheRepository(client), client, container
}

func TestSingleUseTokenCacheRepository_Register(t *testing.T) {
	tests := []struct {
		name          string
		purpose       singleusetoken.Purpose
		jti           string
		value         string
		ttl           time.Duration
		expectError   bool
		validateRedis func(t *testing.T, client *goredis.Client, purpose singleusetoken.Purpose, jti string)
	}{
		{
			name:    "happy path - Register set-password salva auth:set-password:{jti} com TTL",
			purpose: singleusetoken.PurposeSetPassword,
			jti:     "jti-set-pwd",
			value:   "subject-xyz",
			ttl:     24 * time.Hour,
			validateRedis: func(t *testing.T, client *goredis.Client, p singleusetoken.Purpose, jti string) {
				key := "auth:" + string(p) + ":" + jti
				val, err := client.Get(context.Background(), key).Result()
				require.NoError(t, err)
				assert.Equal(t, "subject-xyz", val)

				ttl, err := client.TTL(context.Background(), key).Result()
				require.NoError(t, err)
				assert.InDelta(t, (24 * time.Hour).Seconds(), ttl.Seconds(), 1.0)
			},
		},
		{
			name:    "happy path - Register reset-password salva auth:reset-password:{jti}",
			purpose: singleusetoken.PurposeResetPassword,
			jti:     "jti-reset",
			value:   "subject-reset",
			ttl:     1 * time.Hour,
			validateRedis: func(t *testing.T, client *goredis.Client, p singleusetoken.Purpose, jti string) {
				key := "auth:" + string(p) + ":" + jti
				val, err := client.Get(context.Background(), key).Result()
				require.NoError(t, err)
				assert.Equal(t, "subject-reset", val)
			},
		},
		{
			name:        "failure - jti vazio retorna erro",
			purpose:     singleusetoken.PurposeSetPassword,
			jti:         "",
			value:       "subject",
			ttl:         1 * time.Hour,
			expectError: true,
		},
		{
			name:        "failure - TTL zero retorna erro",
			purpose:     singleusetoken.PurposeSetPassword,
			jti:         "jti-zero-ttl",
			value:       "subject",
			ttl:         0,
			expectError: true,
		},
		{
			name:        "failure - TTL negativo retorna erro",
			purpose:     singleusetoken.PurposeSetPassword,
			jti:         "jti-neg-ttl",
			value:       "subject",
			ttl:         -1 * time.Second,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, client, container := setupRepository(t)
			defer container.Terminate(t)
			defer client.Close()

			err := repo.Register(context.Background(), tt.purpose, tt.jti, tt.value, tt.ttl)

			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.validateRedis != nil {
				tt.validateRedis(t, client, tt.purpose, tt.jti)
			}
		})
	}
}

func TestSingleUseTokenCacheRepository_Consume(t *testing.T) {
	tests := []struct {
		name        string
		purpose     singleusetoken.Purpose
		jti         string
		seed        bool
		expectOK    bool
		expectError bool
		killRedis   bool
	}{
		{
			name:     "happy path - Consume token existente retorna true e deleta",
			purpose:  singleusetoken.PurposeSetPassword,
			jti:      "jti-existing",
			seed:     true,
			expectOK: true,
		},
		{
			name:     "happy path - Consume token inexistente retorna false",
			purpose:  singleusetoken.PurposeSetPassword,
			jti:      "jti-ghost",
			seed:     false,
			expectOK: false,
		},
		{
			name:     "happy path - jti vazio retorna false sem tocar no Redis",
			purpose:  singleusetoken.PurposeSetPassword,
			jti:      "",
			seed:     false,
			expectOK: false,
		},
		{
			name:        "failure - Redis down retorna erro",
			purpose:     singleusetoken.PurposeSetPassword,
			jti:         "jti-x",
			seed:        false,
			killRedis:   true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, client, container := setupRepository(t)
			defer client.Close()

			if tt.seed {
				key := "auth:" + string(tt.purpose) + ":" + tt.jti
				require.NoError(t, client.Set(context.Background(), key, "value", 10*time.Second).Err())
			}

			if tt.killRedis {
				container.Terminate(t)
			} else {
				defer container.Terminate(t)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			got, err := repo.Consume(ctx, tt.purpose, tt.jti)

			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectOK, got)
		})
	}
}
