package event

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithRetry(t *testing.T) {
	tests := []struct {
		name        string
		cfg         RetryConfig
		operation   func() func(context.Context) error
		expectError bool
		errContains string
	}{
		{
			name: "success - returns nil when operation succeeds on first attempt",
			cfg:  RetryConfig{MaxRetries: 3, Delay: 0},
			operation: func() func(context.Context) error {
				return func(ctx context.Context) error {
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "success - returns nil when operation succeeds on second attempt",
			cfg:  RetryConfig{MaxRetries: 3, Delay: 0},
			operation: func() func(context.Context) error {
				callCount := 0
				return func(ctx context.Context) error {
					callCount++
					if callCount < 2 {
						return errors.New("temporary")
					}
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "success - returns nil when operation succeeds on last attempt",
			cfg:  RetryConfig{MaxRetries: 3, Delay: 0},
			operation: func() func(context.Context) error {
				callCount := 0
				return func(ctx context.Context) error {
					callCount++
					if callCount < 3 {
						return errors.New("temporary")
					}
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "failure - returns error after exhausting all retries",
			cfg:  RetryConfig{MaxRetries: 3, Delay: 0},
			operation: func() func(context.Context) error {
				return func(ctx context.Context) error {
					return errors.New("persistent failure")
				}
			},
			expectError: true,
			errContains: "operation failed after 3 retries",
		},
		{
			name: "failure - returns error when context is cancelled during retry",
			cfg:  RetryConfig{MaxRetries: 3, Delay: 5000000000},
			operation: func() func(context.Context) error {
				return func(ctx context.Context) error {
					return errors.New("will retry")
				}
			},
			expectError: true,
			errContains: "retry cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			if tt.errContains == "retry cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			err := WithRetry(ctx, tt.cfg, tt.operation())

			if tt.expectError {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
