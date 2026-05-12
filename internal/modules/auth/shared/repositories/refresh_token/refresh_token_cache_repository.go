package refreshtoken

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const refreshTokenKeyPrefix = "auth:refresh:"

type RefreshTokenPayload struct {
	SessionID string
	SubjectID string
	Role      string
	TTL       time.Duration
}

type RefreshTokenCacheRepository struct {
	client *redis.Client
}

func NewRefreshTokenCacheRepository(client *redis.Client) *RefreshTokenCacheRepository {
	return &RefreshTokenCacheRepository{client: client}
}

type cachedPayload struct {
	SessionID string `json:"session_id"`
	SubjectID string `json:"subject_id"`
	Role      string `json:"role"`
}

func (r *RefreshTokenCacheRepository) Save(ctx context.Context, refreshTokenHash string, payload RefreshTokenPayload) error {
	if payload.TTL <= 0 {
		return fmt.Errorf(ErrInvalidRefreshTokenTTL, payload.TTL)
	}
	body, err := json.Marshal(cachedPayload{
		SessionID: payload.SessionID,
		SubjectID: payload.SubjectID,
		Role:      payload.Role,
	})
	if err != nil {
		return fmt.Errorf(FailedToSaveRefreshTokenErr, err)
	}
	key := refreshTokenKeyPrefix + refreshTokenHash
	if err := r.client.Set(ctx, key, body, payload.TTL).Err(); err != nil {
		return fmt.Errorf(FailedToSaveRefreshTokenErr, err)
	}
	return nil
}

func (r *RefreshTokenCacheRepository) Find(ctx context.Context, refreshTokenHash string) (*RefreshTokenPayload, error) {
	key := refreshTokenKeyPrefix + refreshTokenHash
	raw, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf(FailedToFindRefreshTokenErr, err)
	}
	var cached cachedPayload
	if err := json.Unmarshal([]byte(raw), &cached); err != nil {
		return nil, fmt.Errorf(FailedToFindRefreshTokenErr, err)
	}
	return &RefreshTokenPayload{
		SessionID: cached.SessionID,
		SubjectID: cached.SubjectID,
		Role:      cached.Role,
	}, nil
}

func (r *RefreshTokenCacheRepository) Delete(ctx context.Context, refreshTokenHash string) error {
	key := refreshTokenKeyPrefix + refreshTokenHash
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf(FailedToDeleteRefreshTokenErr, err)
	}
	return nil
}
