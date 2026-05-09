package policy_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func TestAccessPolicy(t *testing.T) {
	tests := []struct {
		name                 string
		build                func() policy.AccessPolicy
		expectedAllowPublic  bool
		expectedRequireOwner bool
		expectedRoles        []role.Role
	}{
		{
			name:                 "PublicAccess retorna AllowPublic=true e AllowedRoles vazio",
			build:                policy.PublicAccess,
			expectedAllowPublic:  true,
			expectedRequireOwner: false,
			expectedRoles:        nil,
		},
		{
			name: "AuthenticatedAccess com roles retorna AllowedRoles corretos e RequireOwnership=false",
			build: func() policy.AccessPolicy {
				return policy.AuthenticatedAccess(role.Specialist, role.Admin)
			},
			expectedAllowPublic:  false,
			expectedRequireOwner: false,
			expectedRoles:        []role.Role{role.Specialist, role.Admin},
		},
		{
			name: "AuthenticatedAccess sem roles expande para todas as roles autenticadas",
			build: func() policy.AccessPolicy {
				return policy.AuthenticatedAccess()
			},
			expectedAllowPublic:  false,
			expectedRequireOwner: false,
			expectedRoles:        []role.Role{role.Specialist, role.Patient, role.Admin},
		},
		{
			name:                 "AnyAuthenticated e alias de AuthenticatedAccess sem argumentos",
			build:                policy.AnyAuthenticated,
			expectedAllowPublic:  false,
			expectedRequireOwner: false,
			expectedRoles:        []role.Role{role.Specialist, role.Patient, role.Admin},
		},
		{
			name: "OwnedAccess retorna AllowedRoles com a role e RequireOwnership=true",
			build: func() policy.AccessPolicy {
				return policy.OwnedAccess(role.Specialist)
			},
			expectedAllowPublic:  false,
			expectedRequireOwner: true,
			expectedRoles:        []role.Role{role.Specialist},
		},
		{
			name:                 "AdminReadOnly retorna AllowedRoles=[admin] e RequireOwnership=false",
			build:                policy.AdminReadOnly,
			expectedAllowPublic:  false,
			expectedRequireOwner: false,
			expectedRoles:        []role.Role{role.Admin},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.build()
			assert.Equal(t, tt.expectedAllowPublic, p.AllowPublic)
			assert.Equal(t, tt.expectedRequireOwner, p.RequireOwnership)
			assert.Equal(t, tt.expectedRoles, p.AllowedRoles)
		})
	}
}

func TestAccessPolicy_ZeroValueIsDenied(t *testing.T) {
	t.Run("zero-value AccessPolicy e bloqueado pelo enforcer com ErrForbidden", func(t *testing.T) {
		var zero policy.AccessPolicy
		enforcer := policy.NewLocalEnforcer()

		err := enforcer.Enforce(context.Background(), zero, validClaimsFactory(), "any")

		require.Error(t, err)
		assert.ErrorIs(t, err, autherrors.ErrForbidden)
	})

	t.Run("zero-value AccessPolicy e bloqueado por EnforceRoleOnly com ErrForbidden", func(t *testing.T) {
		var zero policy.AccessPolicy
		enforcer := policy.NewLocalEnforcer()

		err := enforcer.EnforceRoleOnly(context.Background(), zero, validClaimsFactory())

		require.Error(t, err)
		assert.ErrorIs(t, err, autherrors.ErrForbidden)
	})
}

func TestAccessPolicy_Immutability(t *testing.T) {
	t.Run("mutacao do AllowedRoles retornado nao afeta nova policy criada depois", func(t *testing.T) {
		first := policy.AuthenticatedAccess(role.Specialist, role.Admin)
		first.AllowedRoles[0] = role.Patient

		second := policy.AuthenticatedAccess(role.Specialist, role.Admin)

		assert.Equal(t, []role.Role{role.Specialist, role.Admin}, second.AllowedRoles)
	})

	t.Run("mutacao do slice passado ao construtor nao afeta policy ja construida", func(t *testing.T) {
		roles := []role.Role{role.Specialist, role.Admin}
		p := policy.AuthenticatedAccess(roles...)

		roles[0] = role.Patient

		assert.Equal(t, []role.Role{role.Specialist, role.Admin}, p.AllowedRoles)
	})
}
