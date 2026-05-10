package cache_test

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redistest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/cache/redis"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/validate-token/adapters/outbound/cache"
	platformredis "github.com/lgustavopalmieri/healing-specialist/internal/platform/redis"
)

const blacklistKeyPrefix = "auth:blacklist:"

func setupRepository(t *testing.T) (*cache.BlacklistCacheRepository, *goredis.Client) {
	t.Helper()
	container := redistest.SetupRedisContainer(t)
	t.Cleanup(func() { container.Terminate(t) })

	client, err := platformredis.NewClient(context.Background(), container.Config())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	return cache.NewBlacklistCacheRepository(client), client
}

func TestBlacklistRepository_IsBlacklisted(t *testing.T) {
	tests := []struct {
		name        string
		jti         string
		seedValue   string
		expectFound bool
	}{
		{
			name:        "happy path - jti com valor '1' retorna blacklisted true",
			jti:         "jti-blacklisted",
			seedValue:   "1",
			expectFound: true,
		},
		{
			name:        "happy path - jti com valor diferente de '1' retorna false",
			jti:         "jti-other-value",
			seedValue:   "anything",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, client := setupRepository(t)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			key := blacklistKeyPrefix + tt.jti
			require.NoError(t, client.Set(ctx, key, tt.seedValue, 0).Err())
			t.Cleanup(func() { _ = client.Del(context.Background(), key).Err() })

			got, err := repo.IsBlacklisted(ctx, tt.jti)

			require.NoError(t, err)
			assert.Equal(t, tt.expectFound, got)
		})
	}
}

func TestBlacklistRepository_IsBlacklisted_NotFound(t *testing.T) {
	t.Run("returns false when key does not exist", func(t *testing.T) {
		repo, _ := setupRepository(t)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		got, err := repo.IsBlacklisted(ctx, "jti-missing")

		require.NoError(t, err)
		assert.False(t, got)
	})
}

func TestBlacklistRepository_IsBlacklisted_EmptyJTI(t *testing.T) {
	t.Run("empty jti skips redis and returns false", func(t *testing.T) {
		repo, _ := setupRepository(t)
		ctx := context.Background()

		got, err := repo.IsBlacklisted(ctx, "")

		require.NoError(t, err)
		assert.False(t, got)
	})
}

func TestBlacklistRepository_IsBlacklisted_RedisError(t *testing.T) {
	t.Run("returns wrapped error when redis is unreachable", func(t *testing.T) {
		container := redistest.SetupRedisContainer(t)
		client, err := platformredis.NewClient(context.Background(), container.Config())
		require.NoError(t, err)

		repo := cache.NewBlacklistCacheRepository(client)

		container.Terminate(t)
		_ = client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		got, err := repo.IsBlacklisted(ctx, "jti-any")

		require.Error(t, err)
		assert.False(t, got)
	})
}
