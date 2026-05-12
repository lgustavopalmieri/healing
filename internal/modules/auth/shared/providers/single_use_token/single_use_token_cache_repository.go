package singleusetoken

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Purpose string

const (
	PurposeSetPassword   Purpose = "set-password"
	PurposeResetPassword Purpose = "reset-password"
)

func (p Purpose) keyPrefix() string {
	return "auth:" + string(p) + ":"
}

type SingleUseTokenCacheRepository struct {
	client *redis.Client
}

func NewSingleUseTokenCacheRepository(client *redis.Client) *SingleUseTokenCacheRepository {
	return &SingleUseTokenCacheRepository{client: client}
}

func (r *SingleUseTokenCacheRepository) Register(ctx context.Context, purpose Purpose, jti, value string, ttl time.Duration) error {
	if jti == "" {
		return errors.New("jti must not be empty")
	}
	if ttl <= 0 {
		return fmt.Errorf("invalid single-use token TTL: %s", ttl)
	}
	key := purpose.keyPrefix() + jti
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("register single-use token: %w", err)
	}
	return nil
}

func (r *SingleUseTokenCacheRepository) Consume(ctx context.Context, purpose Purpose, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	key := purpose.keyPrefix() + jti
	val, err := r.client.GetDel(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("consume single-use token: %w", err)
	}
	return val != "", nil
}
