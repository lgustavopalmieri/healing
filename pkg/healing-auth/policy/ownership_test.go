package policy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func TestMustMatchSubject(t *testing.T) {
	tests := []struct {
		name            string
		claims          *claims.Claims
		resourceOwnerID string
		expectedRole    role.Role
		expectError     bool
		expectedErr     error
	}{
		{
			name:            "happy path - claims validas, role correta, subject igual retorna nil",
			claims:          validClaimsFactory(),
			resourceOwnerID: "subject-abc",
			expectedRole:    role.Specialist,
		},
		{
			name:            "failure - claims nil retorna ErrUnauthenticated",
			claims:          nil,
			resourceOwnerID: "subject-abc",
			expectedRole:    role.Specialist,
			expectError:     true,
			expectedErr:     autherrors.ErrUnauthenticated,
		},
		{
			name: "failure - claims invalidas (subject vazio) retorna ErrUnauthenticated",
			claims: validClaimsFactory(func(c *claims.Claims) {
				c.Subject = ""
			}),
			resourceOwnerID: "subject-abc",
			expectedRole:    role.Specialist,
			expectError:     true,
			expectedErr:     autherrors.ErrUnauthenticated,
		},
		{
			name:            "failure - role diferente da esperada retorna ErrForbiddenWrongRole",
			claims:          validClaimsFactory(),
			resourceOwnerID: "subject-abc",
			expectedRole:    role.Admin,
			expectError:     true,
			expectedErr:     autherrors.ErrForbiddenWrongRole,
		},
		{
			name:            "failure - subject diferente do resourceOwnerID retorna ErrForbiddenNotOwner",
			claims:          validClaimsFactory(),
			resourceOwnerID: "outro-subject",
			expectedRole:    role.Specialist,
			expectError:     true,
			expectedErr:     autherrors.ErrForbiddenNotOwner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.MustMatchSubject(tt.claims, tt.resourceOwnerID, tt.expectedRole)
			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestMustMatchSubject_EmptyResourceOwner(t *testing.T) {
	t.Run("claims validas com resourceOwnerID vazio retorna ErrForbiddenNotOwner", func(t *testing.T) {
		err := policy.MustMatchSubject(validClaimsFactory(), "", role.Specialist)

		require.Error(t, err)
		assert.ErrorIs(t, err, autherrors.ErrForbiddenNotOwner)
	})
}
