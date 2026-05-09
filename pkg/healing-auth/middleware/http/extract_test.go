package http_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	authhttp "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/http"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		expected    string
		expectError bool
		expectedErr error
	}{
		{
			name:     "happy path - 'Bearer abc' retorna 'abc'",
			header:   "Bearer abc",
			expected: "abc",
		},
		{
			name:        "failure - header vazio retorna ErrUnauthenticated",
			header:      "",
			expectError: true,
			expectedErr: autherrors.ErrUnauthenticated,
		},
		{
			name:        "failure - header sem prefixo Bearer retorna ErrInvalidToken",
			header:      "Basic abc",
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
		{
			name:        "failure - 'Bearer ' sem token retorna ErrInvalidToken",
			header:      "Bearer ",
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
		{
			name:        "failure - 'Bearer   ' somente espacos retorna ErrInvalidToken",
			header:      "Bearer    ",
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := authhttp.ExtractBearerToken(tt.header)
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
