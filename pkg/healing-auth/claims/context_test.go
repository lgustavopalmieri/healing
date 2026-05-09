package claims_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func validClaimsFactory(overrides ...func(*claims.Claims)) *claims.Claims {
	c := &claims.Claims{
		Subject:  "subject-123",
		Role:     role.Specialist,
		Email:    "specialist@healing.com",
		Provider: provider.Password,
		TokenID:  "token-id-abc",
		IssuedAt: time.Now().Add(-5 * time.Minute),
		ExpireAt: time.Now().Add(55 * time.Minute),
		Issuer:   "healing-specialist",
		Audience: "healing-platform",
	}
	for _, o := range overrides {
		o(c)
	}
	return c
}

func TestWithClaims_FromContext(t *testing.T) {
	t.Run("ciclo ida e volta - claims injetadas sao recuperadas", func(t *testing.T) {
		original := validClaimsFactory()
		ctx := claims.WithClaims(context.Background(), original)

		got, ok := claims.FromContext(ctx)

		require.True(t, ok)
		assert.Equal(t, original.Subject, got.Subject)
		assert.Equal(t, original.Role, got.Role)
		assert.Equal(t, original.TokenID, got.TokenID)
		assert.Equal(t, original.Email, got.Email)
	})
}

func TestFromContext_NoClaims(t *testing.T) {
	t.Run("context sem claims retorna false", func(t *testing.T) {
		got, ok := claims.FromContext(context.Background())

		assert.False(t, ok)
		assert.Nil(t, got)
	})
}

func TestMustFromContext_NoClaims(t *testing.T) {
	t.Run("context sem claims retorna ErrNoClaims", func(t *testing.T) {
		got, err := claims.MustFromContext(context.Background())

		require.Error(t, err)
		assert.ErrorIs(t, err, autherrors.ErrNoClaims)
		assert.Nil(t, got)
	})
}

func TestMustFromContext_InvalidClaims(t *testing.T) {
	tests := []struct {
		name     string
		override func(*claims.Claims)
	}{
		{
			name:     "claims com Subject vazio retorna ErrInvalidClaims",
			override: func(c *claims.Claims) { c.Subject = "" },
		},
		{
			name:     "claims com Role invalida retorna ErrInvalidClaims",
			override: func(c *claims.Claims) { c.Role = role.Role("unknown") },
		},
		{
			name:     "claims com TokenID vazio retorna ErrInvalidClaims",
			override: func(c *claims.Claims) { c.TokenID = "" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validClaimsFactory(tt.override)
			ctx := claims.WithClaims(context.Background(), c)

			got, err := claims.MustFromContext(ctx)

			require.Error(t, err)
			assert.ErrorIs(t, err, autherrors.ErrInvalidClaims)
			assert.Nil(t, got)
		})
	}
}

func TestClaims_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		expireAt time.Time
		now      time.Time
		expected bool
	}{
		{
			name:     "token com ExpireAt no passado retorna true",
			expireAt: time.Now().Add(-1 * time.Hour),
			now:      time.Now(),
			expected: true,
		},
		{
			name:     "token com ExpireAt no futuro retorna false",
			expireAt: time.Now().Add(1 * time.Hour),
			now:      time.Now(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validClaimsFactory(func(c *claims.Claims) {
				c.ExpireAt = tt.expireAt
			})
			assert.Equal(t, tt.expected, c.IsExpired(tt.now))
		})
	}
}

func TestClaims_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    *claims.Claims
		expected bool
	}{
		{
			name:     "claims validas retorna true",
			input:    validClaimsFactory(),
			expected: true,
		},
		{
			name:     "nil retorna false",
			input:    nil,
			expected: false,
		},
		{
			name:     "subject vazio retorna false",
			input:    validClaimsFactory(func(c *claims.Claims) { c.Subject = "" }),
			expected: false,
		},
		{
			name:     "role invalida retorna false",
			input:    validClaimsFactory(func(c *claims.Claims) { c.Role = role.Role("bad") }),
			expected: false,
		},
		{
			name:     "tokenID vazio retorna false",
			input:    validClaimsFactory(func(c *claims.Claims) { c.TokenID = "" }),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.Valid())
		})
	}
}
