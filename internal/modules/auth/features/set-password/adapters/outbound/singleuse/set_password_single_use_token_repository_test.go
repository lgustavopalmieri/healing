package singleuse_test

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redistest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/cache/redis"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/adapters/outbound/singleuse"
	singleusetoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/providers/single_use_token"
	platformredis "github.com/lgustavopalmieri/healing-specialist/internal/platform/redis"
)

func setupRepository(t *testing.T) (*singleuse.SetPasswordSingleUseTokenRepository, *goredis.Client) {
	t.Helper()
	container := redistest.SetupRedisContainer(t)
	t.Cleanup(func() { container.Terminate(t) })

	client, err := platformredis.NewClient(context.Background(), container.Config())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	store := singleusetoken.NewSingleUseTokenCacheRepository(client)
	return singleuse.NewSetPasswordSingleUseTokenRepository(store), client
}

func TestSetPasswordSingleUseTokenRepository_Consume(t *testing.T) {
	tests := []struct {
		name          string
		jti           string
		seed          func(t *testing.T, client *goredis.Client, jti string)
		expectOK      bool
		expectError   bool
		validateRedis func(t *testing.T, client *goredis.Client, jti string)
	}{
		{
			name: "happy path - jti presente retorna true e deleta a chave (GETDEL atomico)",
			jti:  "jti-1",
			seed: func(t *testing.T, client *goredis.Client, jti string) {
				key := "auth:set-password:" + jti
				require.NoError(t, client.Set(context.Background(), key, "subject-value", 10*time.Second).Err())
			},
			expectOK: true,
			validateRedis: func(t *testing.T, client *goredis.Client, jti string) {
				_, err := client.Get(context.Background(), "auth:set-password:"+jti).Result()
				assert.ErrorIs(t, err, goredis.Nil)
			},
		},
		{
			name:     "happy path - jti ausente retorna false sem erro",
			jti:      "jti-missing",
			expectOK: false,
		},
		{
			name: "happy path - segundo Consume do mesmo jti retorna false (ja foi deletado)",
			jti:  "jti-2",
			seed: func(t *testing.T, client *goredis.Client, jti string) {
				key := "auth:set-password:" + jti
				require.NoError(t, client.Set(context.Background(), key, "subject-value", 10*time.Second).Err())
			},
			expectOK: false,
		},
		{
			name:     "happy path - jti vazio retorna false sem consultar Redis",
			jti:      "",
			expectOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, client := setupRepository(t)

			if tt.seed != nil {
				tt.seed(t, client, tt.jti)
			}

			if tt.name == "happy path - segundo Consume do mesmo jti retorna false (ja foi deletado)" {
				// primeiro consume pra esvaziar
				_, err := repo.Consume(context.Background(), tt.jti)
				require.NoError(t, err)
			}

			got, err := repo.Consume(context.Background(), tt.jti)

			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectOK, got)

			if tt.validateRedis != nil {
				tt.validateRedis(t, client, tt.jti)
			}
		})
	}
}
