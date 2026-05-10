package admin_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/admin"
)

func adminFactory(overrides ...func(*admin.NewAdminInput)) admin.NewAdminInput {
	in := admin.NewAdminInput{
		Name:    "Platform Admin",
		Email:   "admin@healing.local",
		SubRole: admin.SubRoleAdmin,
	}
	for _, o := range overrides {
		o(&in)
	}
	return in
}

func TestAdmin_New(t *testing.T) {
	tests := []struct {
		name        string
		input       admin.NewAdminInput
		expectError bool
		expectedErr error
	}{
		{
			name:  "happy path - sub-role admin cria admin active",
			input: adminFactory(),
		},
		{
			name:  "happy path - sub-role support cria admin active",
			input: adminFactory(func(in *admin.NewAdminInput) { in.SubRole = admin.SubRoleSupport }),
		},
		{
			name:  "happy path - sub-role moderator cria admin active",
			input: adminFactory(func(in *admin.NewAdminInput) { in.SubRole = admin.SubRoleModerator }),
		},
		{
			name: "failure - sub-role invalido retorna ErrInvalidSubRole",
			input: adminFactory(func(in *admin.NewAdminInput) {
				in.SubRole = admin.SubRole("owner")
			}),
			expectError: true,
			expectedErr: admin.ErrInvalidSubRole,
		},
		{
			name: "failure - sub-role vazio retorna ErrInvalidSubRole",
			input: adminFactory(func(in *admin.NewAdminInput) {
				in.SubRole = admin.SubRole("")
			}),
			expectError: true,
			expectedErr: admin.ErrInvalidSubRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := admin.NewAdmin(tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)

			_, uerr := uuid.Parse(got.ID)
			assert.NoError(t, uerr, "ID should be a valid UUID")

			assert.Equal(t, tt.input.Name, got.Name)
			assert.Equal(t, tt.input.Email, got.Email)
			assert.Equal(t, tt.input.SubRole, got.SubRole)
			assert.Equal(t, admin.StatusActive, got.Status)
			assert.False(t, got.CreatedAt.IsZero())
			assert.False(t, got.UpdatedAt.IsZero())
		})
	}
}

func TestAdmin_Activate(t *testing.T) {
	t.Run("happy path - inactive -> active", func(t *testing.T) {
		a, err := admin.NewAdmin(adminFactory())
		require.NoError(t, err)
		a.Status = admin.StatusInactive
		previousUpdatedAt := a.UpdatedAt

		time.Sleep(time.Millisecond)
		err = a.Activate()

		require.NoError(t, err)
		assert.Equal(t, admin.StatusActive, a.Status)
		assert.True(t, a.UpdatedAt.After(previousUpdatedAt))
	})

	t.Run("failure - active -> active retorna ErrInvalidStatusTransition", func(t *testing.T) {
		a, err := admin.NewAdmin(adminFactory())
		require.NoError(t, err)

		err = a.Activate()

		require.Error(t, err)
		assert.ErrorIs(t, err, admin.ErrInvalidStatusTransition)
	})
}

func TestAdmin_Deactivate(t *testing.T) {
	t.Run("happy path - active -> inactive", func(t *testing.T) {
		a, err := admin.NewAdmin(adminFactory())
		require.NoError(t, err)

		err = a.Deactivate()

		require.NoError(t, err)
		assert.Equal(t, admin.StatusInactive, a.Status)
	})

	t.Run("failure - inactive -> inactive retorna ErrInvalidStatusTransition", func(t *testing.T) {
		a, err := admin.NewAdmin(adminFactory())
		require.NoError(t, err)
		require.NoError(t, a.Deactivate())

		err = a.Deactivate()

		require.Error(t, err)
		assert.ErrorIs(t, err, admin.ErrInvalidStatusTransition)
	})
}

func TestSubRole_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    admin.SubRole
		expected bool
	}{
		{name: "true para admin", input: admin.SubRoleAdmin, expected: true},
		{name: "true para support", input: admin.SubRoleSupport, expected: true},
		{name: "true para moderator", input: admin.SubRoleModerator, expected: true},
		{name: "false para string desconhecida", input: admin.SubRole("owner"), expected: false},
		{name: "false para string vazia", input: admin.SubRole(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.Valid())
		})
	}
}

func TestSubRole_Permissions(t *testing.T) {
	tests := []struct {
		name               string
		sub                admin.SubRole
		canReadSpecialists bool
		canReadPatients    bool
		canBlockAccounts   bool
		canViewAuditLogs   bool
		canModerateReviews bool
	}{
		{
			name:               "admin pode tudo",
			sub:                admin.SubRoleAdmin,
			canReadSpecialists: true,
			canReadPatients:    true,
			canBlockAccounts:   true,
			canViewAuditLogs:   true,
			canModerateReviews: true,
		},
		{
			name:               "support le specialists e patients, ve audit, mas nao bloqueia nem modera",
			sub:                admin.SubRoleSupport,
			canReadSpecialists: true,
			canReadPatients:    true,
			canBlockAccounts:   false,
			canViewAuditLogs:   true,
			canModerateReviews: false,
		},
		{
			name:               "moderator le specialists e modera reviews, nada mais",
			sub:                admin.SubRoleModerator,
			canReadSpecialists: true,
			canReadPatients:    false,
			canBlockAccounts:   false,
			canViewAuditLogs:   false,
			canModerateReviews: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.canReadSpecialists, tt.sub.CanReadSpecialists())
			assert.Equal(t, tt.canReadPatients, tt.sub.CanReadPatients())
			assert.Equal(t, tt.canBlockAccounts, tt.sub.CanBlockAccounts())
			assert.Equal(t, tt.canViewAuditLogs, tt.sub.CanViewAuditLogs())
			assert.Equal(t, tt.canModerateReviews, tt.sub.CanModerateReviews())
		})
	}
}

func TestStatus_Valid(t *testing.T) {
	assert.True(t, admin.StatusActive.Valid())
	assert.True(t, admin.StatusInactive.Valid())
	assert.False(t, admin.Status("suspended").Valid())
	assert.False(t, admin.Status("").Valid())
}

func TestStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     admin.Status
		to       admin.Status
		expected bool
	}{
		{name: "active -> inactive", from: admin.StatusActive, to: admin.StatusInactive, expected: true},
		{name: "active -> active", from: admin.StatusActive, to: admin.StatusActive, expected: false},
		{name: "inactive -> active", from: admin.StatusInactive, to: admin.StatusActive, expected: true},
		{name: "inactive -> inactive", from: admin.StatusInactive, to: admin.StatusInactive, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.from.CanTransitionTo(tt.to))
		})
	}
}
