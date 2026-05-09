package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	platformredis "github.com/lgustavopalmieri/healing-specialist/internal/platform/redis"
)

func InitRedis(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	log.Printf("Connecting to Redis (%s:%d)...", cfg.Redis.Host, cfg.Redis.Port)

	client, err := platformredis.NewClient(ctx, platformredis.Config{
		Host:         cfg.Redis.Host,
		Port:         cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize redis: %w", err)
	}

	log.Printf("Redis connected (db=%d, pool=%d/%d)",
		cfg.Redis.DB,
		cfg.Redis.MinIdleConns,
		cfg.Redis.PoolSize,
	)

	return client, nil
}
