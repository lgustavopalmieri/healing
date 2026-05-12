package blacklist

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const blacklistKeyPrefix = "auth:blacklist:"

type BlacklistCacheRepository struct {
	client *redis.Client
}

func NewBlacklistCacheRepository(client *redis.Client) *BlacklistCacheRepository {
	return &BlacklistCacheRepository{client: client}
}

func (r *BlacklistCacheRepository) Blacklist(ctx context.Context, jti string, ttl time.Duration) error {
	if jti == "" {
		return nil
	}
	if ttl <= 0 {
		return nil
	}
	key := blacklistKeyPrefix + jti
	if err := r.client.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis set blacklist: %w", err)
	}
	return nil
}

func (r *BlacklistCacheRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	key := blacklistKeyPrefix + jti
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("redis get blacklist: %w", err)
	}
	return val == "1", nil
}
