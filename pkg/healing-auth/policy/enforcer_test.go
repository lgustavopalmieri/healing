package policy_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func TestLocalEnforcer_Enforce(t *testing.T) {
	enforcer := policy.NewLocalEnforcer()
	ctx := context.Background()

	tests := []struct {
		name            string
		accessPolicy    policy.AccessPolicy
		claims          *claims.Claims
		resourceOwnerID string
		expectError     bool
		expectedErr     error
	}{
		{
			name:            "public + sem claims retorna nil",
			accessPolicy:    policy.PublicAccess(),
			claims:          nil,
			resourceOwnerID: "",
		},
		{
			name:            "public + com claims retorna nil",
			accessPolicy:    policy.PublicAccess(),
			claims:          validClaimsFactory(),
			resourceOwnerID: "qualquer",
		},
		{
			name:            "zero-value AccessPolicy retorna ErrForbidden (fail-closed)",
			accessPolicy:    policy.AccessPolicy{},
			claims:          validClaimsFactory(),
			resourceOwnerID: "subject-abc",
			expectError:     true,
			expectedErr:     autherrors.ErrForbidden,
		},
		{
			name:            "nao public + claims nil retorna ErrUnauthenticated",
			accessPolicy:    policy.AuthenticatedAccess(role.Specialist),
			claims:          nil,
			resourceOwnerID: "",
			expectError:     true,
			expectedErr:     autherrors.ErrUnauthenticated,
		},
		{
			name:         "nao public + claims invalidas retorna ErrUnauthenticated",
			accessPolicy: policy.AuthenticatedAccess(role.Specialist),
			claims: validClaimsFactory(func(c *claims.Claims) {
				c.Subject = ""
			}),
			resourceOwnerID: "",
			expectError:     true,
			expectedErr:     autherrors.ErrUnauthenticated,
		},
		{
			name:            "nao public + role errada retorna ErrForbiddenWrongRole",
			accessPolicy:    policy.AuthenticatedAccess(role.Admin),
			claims:          validClaimsFactory(),
			resourceOwnerID: "",
			expectError:     true,
			expectedErr:     autherrors.ErrForbiddenWrongRole,
		},
		{
			name:            "owned + subject igual retorna nil",
			accessPolicy:    policy.OwnedAccess(role.Specialist),
			claims:          validClaimsFactory(),
			resourceOwnerID: "subject-abc",
		},
		{
			name:            "owned + subject diferente retorna ErrForbiddenNotOwner",
			accessPolicy:    policy.OwnedAccess(role.Specialist),
			claims:          validClaimsFactory(),
			resourceOwnerID: "outro-subject",
			expectError:     true,
			expectedErr:     autherrors.ErrForbiddenNotOwner,
		},
		{
			name:            "owned + role errada retorna ErrForbiddenWrongRole (nao ErrForbiddenNotOwner)",
			accessPolicy:    policy.OwnedAccess(role.Specialist),
			claims:          validClaimsFactory(func(c *claims.Claims) { c.Role = role.Admin }),
			resourceOwnerID: "subject-abc",
			expectError:     true,
			expectedErr:     autherrors.ErrForbiddenWrongRole,
		},
		{
			name:            "AuthenticatedAccess + role permitida retorna nil",
			accessPolicy:    policy.AuthenticatedAccess(role.Specialist, role.Admin),
			claims:          validClaimsFactory(),
			resourceOwnerID: "",
		},
		{
			name:            "AuthenticatedAccess com multiplas roles - specialist passa",
			accessPolicy:    policy.AuthenticatedAccess(role.Specialist, role.Patient, role.Admin),
			claims:          validClaimsFactory(),
			resourceOwnerID: "",
		},
		{
			name:            "AuthenticatedAccess com multiplas roles - patient passa",
			accessPolicy:    policy.AuthenticatedAccess(role.Specialist, role.Patient, role.Admin),
			claims:          validClaimsFactory(func(c *claims.Claims) { c.Role = role.Patient }),
			resourceOwnerID: "",
		},
		{
			name:            "AuthenticatedAccess com multiplas roles - admin passa",
			accessPolicy:    policy.AuthenticatedAccess(role.Specialist, role.Patient, role.Admin),
			claims:          validClaimsFactory(func(c *claims.Claims) { c.Role = role.Admin }),
			resourceOwnerID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := enforcer.Enforce(ctx, tt.accessPolicy, tt.claims, tt.resourceOwnerID)
			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestLocalEnforcer_EnforceRoleOnly(t *testing.T) {
	enforcer := policy.NewLocalEnforcer()
	ctx := context.Background()

	tests := []struct {
		name         string
		accessPolicy policy.AccessPolicy
		claims       *claims.Claims
		expectError  bool
		expectedErr  error
	}{
		{
			name:         "public retorna nil sem claims",
			accessPolicy: policy.PublicAccess(),
			claims:       nil,
		},
		{
			name:         "zero-value AccessPolicy retorna ErrForbidden",
			accessPolicy: policy.AccessPolicy{},
			claims:       validClaimsFactory(),
			expectError:  true,
			expectedErr:  autherrors.ErrForbidden,
		},
		{
			name:         "nao public + claims nil retorna ErrUnauthenticated",
			accessPolicy: policy.AuthenticatedAccess(role.Specialist),
			claims:       nil,
			expectError:  true,
			expectedErr:  autherrors.ErrUnauthenticated,
		},
		{
			name:         "role correta retorna nil",
			accessPolicy: policy.AuthenticatedAccess(role.Specialist),
			claims:       validClaimsFactory(),
		},
		{
			name:         "role errada retorna ErrForbiddenWrongRole",
			accessPolicy: policy.AuthenticatedAccess(role.Patient),
			claims:       validClaimsFactory(),
			expectError:  true,
			expectedErr:  autherrors.ErrForbiddenWrongRole,
		},
		{
			name:         "owned policy + role correta retorna nil (sem checar subject)",
			accessPolicy: policy.OwnedAccess(role.Specialist),
			claims:       validClaimsFactory(),
		},
		{
			name:         "owned policy + role correta + claims.Subject vazio ignora ownership",
			accessPolicy: policy.OwnedAccess(role.Specialist),
			claims: validClaimsFactory(func(c *claims.Claims) {
				c.Subject = "qualquer-subject-diferente"
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := enforcer.EnforceRoleOnly(ctx, tt.accessPolicy, tt.claims)
			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
