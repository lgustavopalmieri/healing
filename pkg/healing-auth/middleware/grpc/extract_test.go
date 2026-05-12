package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	authgrpc "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/grpc"
)

func TestExtractBearerToken_GRPC(t *testing.T) {
	tests := []struct {
		name        string
		buildCtx    func() context.Context
		expected    string
		expectError bool
		expectedErr error
	}{
		{
			name: "happy path - metadata com 'authorization: Bearer abc' retorna 'abc'",
			buildCtx: func() context.Context {
				md := metadata.Pairs(authgrpc.AuthorizationMetadataKey, "Bearer abc")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expected: "abc",
		},
		{
			name: "failure - sem metadata no context retorna ErrUnauthenticated",
			buildCtx: func() context.Context {
				return context.Background()
			},
			expectError: true,
			expectedErr: autherrors.ErrUnauthenticated,
		},
		{
			name: "failure - metadata sem authorization retorna ErrUnauthenticated",
			buildCtx: func() context.Context {
				md := metadata.Pairs("x-trace-id", "trace-1")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectError: true,
			expectedErr: autherrors.ErrUnauthenticated,
		},
		{
			name: "failure - authorization sem prefixo Bearer retorna ErrInvalidToken",
			buildCtx: func() context.Context {
				md := metadata.Pairs(authgrpc.AuthorizationMetadataKey, "Basic abc")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
		{
			name: "failure - 'Bearer ' sem token retorna ErrInvalidToken",
			buildCtx: func() context.Context {
				md := metadata.Pairs(authgrpc.AuthorizationMetadataKey, "Bearer ")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := authgrpc.ExtractBearerToken(tt.buildCtx())
			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Empty(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
