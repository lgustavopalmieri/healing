package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const blacklistKeyPrefix = "auth:blacklist:"

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
