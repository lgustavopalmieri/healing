package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix = "auth:login-attempts:"
	ttl       = 15 * time.Minute
)

type LoginAttemptsTracker struct {
	client *redis.Client
}

func NewLoginAttemptsTracker(client *redis.Client) *LoginAttemptsTracker {
	return &LoginAttemptsTracker{client: client}
}

func (t *LoginAttemptsTracker) Increment(ctx context.Context, email string) error {
	key := keyPrefix + email
	pipe := t.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("increment login attempts: %w", err)
	}
	return nil
}

func (t *LoginAttemptsTracker) Reset(ctx context.Context, email string) error {
	key := keyPrefix + email
	if err := t.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("reset login attempts: %w", err)
	}
	return nil
}
