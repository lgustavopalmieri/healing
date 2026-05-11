package refreshtoken_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redistest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/cache/redis"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	platformredis "github.com/lgustavopalmieri/healing-specialist/internal/platform/redis"
)

func setupRepository(t *testing.T) (*refreshtoken.RefreshTokenCacheRepository, *goredis.Client, *redistest.RedisContainer) {
	t.Helper()
	container := redistest.SetupRedisContainer(t)
	client, err := platformredis.NewClient(context.Background(), container.Config())
	require.NoError(t, err)
	return refreshtoken.NewRefreshTokenCacheRepository(client), client, container
}

func TestRefreshTokenCacheRepository_Save(t *testing.T) {
	tests := []struct {
		name          string
		hash          string
		payload       refreshtoken.RefreshTokenPayload
		expectError   bool
		validateRedis func(t *testing.T, client *goredis.Client, hash string, expected refreshtoken.RefreshTokenPayload)
	}{
		{
			name: "happy path - salva payload com TTL e chave auth:refresh:{hash}",
			hash: "hash-abc",
			payload: refreshtoken.RefreshTokenPayload{
				SessionID: "sess-1",
				SubjectID: "subject-1",
				Role:      "specialist",
				TTL:       168 * time.Hour,
			},
			validateRedis: func(t *testing.T, client *goredis.Client, hash string, expected refreshtoken.RefreshTokenPayload) {
				raw, err := client.Get(context.Background(), "auth:refresh:"+hash).Result()
				require.NoError(t, err)

				var decoded map[string]string
				require.NoError(t, json.Unmarshal([]byte(raw), &decoded))
				assert.Equal(t, expected.SessionID, decoded["session_id"])
				assert.Equal(t, expected.SubjectID, decoded["subject_id"])
				assert.Equal(t, expected.Role, decoded["role"])

				ttl, err := client.TTL(context.Background(), "auth:refresh:"+hash).Result()
				require.NoError(t, err)
				assert.InDelta(t, expected.TTL.Seconds(), ttl.Seconds(), 1.0)
			},
		},
		{
			name: "failure - TTL zero retorna erro",
			hash: "hash-zero-ttl",
			payload: refreshtoken.RefreshTokenPayload{
				SessionID: "sess-1",
				SubjectID: "subject-1",
				Role:      "specialist",
				TTL:       0,
			},
			expectError: true,
		},
		{
			name: "failure - TTL negativo retorna erro",
			hash: "hash-neg-ttl",
			payload: refreshtoken.RefreshTokenPayload{
				SessionID: "sess-1",
				SubjectID: "subject-1",
				Role:      "specialist",
				TTL:       -1 * time.Hour,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, client, container := setupRepository(t)
			defer container.Terminate(t)
			defer client.Close()

			err := repo.Save(context.Background(), tt.hash, tt.payload)

			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.validateRedis != nil {
				tt.validateRedis(t, client, tt.hash, tt.payload)
			}
		})
	}
}

func TestRefreshTokenCacheRepository_Find(t *testing.T) {
	tests := []struct {
		name        string
		hash        string
		seed        bool
		expectNil   bool
		expectError bool
		killRedis   bool
	}{
		{
			name: "happy path - retorna payload deserializado",
			hash: "hash-xyz",
			seed: true,
		},
		{
			name:      "happy path - retorna nil sem erro quando key nao existe",
			hash:      "hash-missing",
			seed:      false,
			expectNil: true,
		},
		{
			name:        "failure - Redis down retorna erro",
			hash:        "hash-x",
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
				err := repo.Save(context.Background(), tt.hash, refreshtoken.RefreshTokenPayload{
					SessionID: "sess-seeded",
					SubjectID: "subject-seeded",
					Role:      "specialist",
					TTL:       10 * time.Second,
				})
				require.NoError(t, err)
			}

			if tt.killRedis {
				container.Terminate(t)
			} else {
				defer container.Terminate(t)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			got, err := repo.Find(ctx, tt.hash)

			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.expectNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, "sess-seeded", got.SessionID)
			assert.Equal(t, "subject-seeded", got.SubjectID)
			assert.Equal(t, "specialist", got.Role)
		})
	}
}

func TestRefreshTokenCacheRepository_Delete(t *testing.T) {
	tests := []struct {
		name string
		hash string
		seed bool
	}{
		{
			name: "happy path - remove chave existente",
			hash: "hash-del",
			seed: true,
		},
		{
			name: "happy path - remover chave inexistente nao retorna erro",
			hash: "hash-no-exist",
			seed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, client, container := setupRepository(t)
			defer container.Terminate(t)
			defer client.Close()

			if tt.seed {
				err := repo.Save(context.Background(), tt.hash, refreshtoken.RefreshTokenPayload{
					SessionID: "x",
					SubjectID: "y",
					Role:      "specialist",
					TTL:       10 * time.Second,
				})
				require.NoError(t, err)
			}

			err := repo.Delete(context.Background(), tt.hash)
			require.NoError(t, err)

			_, err = client.Get(context.Background(), "auth:refresh:"+tt.hash).Result()
			assert.ErrorIs(t, err, goredis.Nil)
		})
	}
}
