package redis

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	platformredis "github.com/lgustavopalmieri/healing-specialist/internal/platform/redis"
)

type RedisContainer struct {
	Container testcontainers.Container
	Host      string
	Port      int
}

func SetupRedisContainer(t *testing.T) *RedisContainer {
	t.Helper()
	ctx := context.Background()

	container, err := tcredis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	mapped, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	port, err := strconv.Atoi(mapped.Port())
	require.NoError(t, err)

	return &RedisContainer{
		Container: container,
		Host:      host,
		Port:      port,
	}
}

func (c *RedisContainer) Config() platformredis.Config {
	return platformredis.Config{
		Host:         c.Host,
		Port:         c.Port,
		Password:     "",
		DB:           0,
		PoolSize:     5,
		MinIdleConns: 1,
	}
}

func (c *RedisContainer) Terminate(t *testing.T) {
	t.Helper()
	if c == nil || c.Container == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = c.Container.Terminate(ctx)
}
