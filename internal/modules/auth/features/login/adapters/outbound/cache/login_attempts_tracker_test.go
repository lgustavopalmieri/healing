package cache_test

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redistest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/cache/redis"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/adapters/outbound/cache"
	platformredis "github.com/lgustavopalmieri/healing-specialist/internal/platform/redis"
)

const keyPrefix = "auth:login-attempts:"

func setupTracker(t *testing.T) (*cache.LoginAttemptsTracker, *goredis.Client) {
	t.Helper()
	container := redistest.SetupRedisContainer(t)
	t.Cleanup(func() { container.Terminate(t) })

	client, err := platformredis.NewClient(context.Background(), container.Config())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	return cache.NewLoginAttemptsTracker(client), client
}

func TestLoginAttemptsTracker_Increment(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		callCount     int
		expectedValue string
		validateTTL   bool
	}{
		{
			name:          "happy path - Increment cria chave com valor 1 e TTL 15min",
			email:         "user@healing.com",
			callCount:     1,
			expectedValue: "1",
			validateTTL:   true,
		},
		{
			name:          "happy path - segundo Increment incrementa pra 2",
			email:         "user2@healing.com",
			callCount:     2,
			expectedValue: "2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker, client := setupTracker(t)

			for i := 0; i < tt.callCount; i++ {
				err := tracker.Increment(context.Background(), tt.email)
				require.NoError(t, err)
			}

			key := keyPrefix + tt.email
			val, err := client.Get(context.Background(), key).Result()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedValue, val)

			if tt.validateTTL {
				ttl, err := client.TTL(context.Background(), key).Result()
				require.NoError(t, err)
				assert.InDelta(t, (15 * time.Minute).Seconds(), ttl.Seconds(), 2.0)
			}
		})
	}
}

func TestLoginAttemptsTracker_Reset(t *testing.T) {
	tests := []struct {
		name  string
		email string
		seed  bool
	}{
		{
			name:  "happy path - Reset remove a chave",
			email: "user@healing.com",
			seed:  true,
		},
		{
			name:  "happy path - Reset em chave inexistente nao retorna erro",
			email: "ghost@healing.com",
			seed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker, client := setupTracker(t)

			if tt.seed {
				require.NoError(t, tracker.Increment(context.Background(), tt.email))
			}

			err := tracker.Reset(context.Background(), tt.email)
			require.NoError(t, err)

			_, err = client.Get(context.Background(), keyPrefix+tt.email).Result()
			assert.ErrorIs(t, err, goredis.Nil)
		})
	}
}
