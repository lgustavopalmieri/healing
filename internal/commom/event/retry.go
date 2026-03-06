package event

import (
	"context"
	"fmt"
	"time"
)

const (
	DefaultMaxRetries = 3
	DefaultRetryDelay = 500 * time.Millisecond
)

type RetryConfig struct {
	MaxRetries int
	Delay      time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: DefaultMaxRetries,
		Delay:      DefaultRetryDelay,
	}
}

func WithRetry(ctx context.Context, cfg RetryConfig, operation func(ctx context.Context) error) error {
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		lastErr = operation(ctx)
		if lastErr == nil {
			return nil
		}

		if attempt < cfg.MaxRetries {
			select {
			case <-ctx.Done():
				return fmt.Errorf("retry cancelled: %w", ctx.Err())
			case <-time.After(cfg.Delay):
			}
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", cfg.MaxRetries, lastErr)
}
