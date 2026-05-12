package blacklist_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redistest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/cache/redis"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/blacklist"
	platformredis "github.com/lgustavopalmieri/healing-specialist/internal/platform/redis"
)

func setupRepository(t *testing.T) *blacklist.BlacklistCacheRepository {
	t.Helper()
	container := redistest.SetupRedisContainer(t)
	t.Cleanup(func() { container.Terminate(t) })

	client, err := platformredis.NewClient(context.Background(), container.Config())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	return blacklist.NewBlacklistCacheRepository(client)
}

func TestBlacklistCacheRepository_Blacklist(t *testing.T) {
	tests := []struct {
		name      string
		jti       string
		ttl       time.Duration
		expectSet bool
	}{
		{
			name:      "happy path - SET auth:blacklist:{jti} com TTL",
			jti:       "jti-to-blacklist",
			ttl:       10 * time.Second,
			expectSet: true,
		},
		{
			name: "happy path - jti vazio nao faz nada",
			jti:  "",
			ttl:  10 * time.Second,
		},
		{
			name: "happy path - TTL zero nao faz nada",
			jti:  "jti-zero-ttl",
			ttl:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupRepository(t)

			err := repo.Blacklist(context.Background(), tt.jti, tt.ttl)
			require.NoError(t, err)

			if tt.expectSet {
				blacklisted, err := repo.IsBlacklisted(context.Background(), tt.jti)
				require.NoError(t, err)
				assert.True(t, blacklisted)
			}
		})
	}
}

func TestBlacklistCacheRepository_IsBlacklisted(t *testing.T) {
	tests := []struct {
		name     string
		jti      string
		seed     bool
		expected bool
	}{
		{
			name:     "happy path - jti blacklisted retorna true",
			jti:      "jti-exists",
			seed:     true,
			expected: true,
		},
		{
			name:     "happy path - jti nao blacklisted retorna false",
			jti:      "jti-missing",
			seed:     false,
			expected: false,
		},
		{
			name:     "happy path - jti vazio retorna false",
			jti:      "",
			seed:     false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupRepository(t)

			if tt.seed {
				require.NoError(t, repo.Blacklist(context.Background(), tt.jti, 10*time.Second))
			}

			got, err := repo.IsBlacklisted(context.Background(), tt.jti)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
