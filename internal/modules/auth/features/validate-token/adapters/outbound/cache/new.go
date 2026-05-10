package cache

import (
	"github.com/redis/go-redis/v9"
)

type BlacklistCacheRepository struct {
	client *redis.Client
}

func NewBlacklistCacheRepository(client *redis.Client) *BlacklistCacheRepository {
	return &BlacklistCacheRepository{client: client}
}
