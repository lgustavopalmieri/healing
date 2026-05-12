package role_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func TestParse_Role(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    role.Role
		expectError bool
		expectedErr error
	}{
		{
			name:     "happy path - specialist retorna Role specialist",
			input:    "specialist",
			expected: role.Specialist,
		},
		{
			name:     "happy path - patient retorna Role patient",
			input:    "patient",
			expected: role.Patient,
		},
		{
			name:     "happy path - admin retorna Role admin",
			input:    "admin",
			expected: role.Admin,
		},
		{
			name:     "happy path - anonymous retorna Role anonymous",
			input:    "anonymous",
			expected: role.Anonymous,
		},
		{
			name:        "failure - string invalida retorna ErrInvalidRole",
			input:       "superadmin",
			expectError: true,
			expectedErr: role.ErrInvalidRole,
		},
		{
			name:        "failure - string vazia retorna ErrInvalidRole",
			input:       "",
			expectError: true,
			expectedErr: role.ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := role.Parse(tt.input)
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

func TestRole_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    role.Role
		expected bool
	}{
		{
			name:     "true para specialist",
			input:    role.Specialist,
			expected: true,
		},
		{
			name:     "true para patient",
			input:    role.Patient,
			expected: true,
		},
		{
			name:     "true para admin",
			input:    role.Admin,
			expected: true,
		},
		{
			name:     "true para anonymous",
			input:    role.Anonymous,
			expected: true,
		},
		{
			name:     "false para string desconhecida",
			input:    role.Role("foo"),
			expected: false,
		},
		{
			name:     "false para string vazia",
			input:    role.Role(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.Valid())
		})
	}
}

func TestRole_Helpers(t *testing.T) {
	assert.True(t, role.Specialist.IsSpecialist())
	assert.False(t, role.Specialist.IsAdmin())
	assert.True(t, role.Patient.IsPatient())
	assert.True(t, role.Admin.IsAdmin())
	assert.True(t, role.Anonymous.IsAnonymous())
	assert.Equal(t, "specialist", role.Specialist.String())
}

func TestMustBeIn(t *testing.T) {
	tests := []struct {
		name        string
		actual      role.Role
		allowed     []role.Role
		expectError bool
		expectedErr error
	}{
		{
			name:    "role presente na lista retorna nil",
			actual:  role.Specialist,
			allowed: []role.Role{role.Specialist, role.Admin},
		},
		{
			name:        "role ausente retorna ErrForbiddenWrongRole",
			actual:      role.Patient,
			allowed:     []role.Role{role.Specialist, role.Admin},
			expectError: true,
			expectedErr: autherrors.ErrForbiddenWrongRole,
		},
		{
			name:    "allowed vazio retorna nil (sem restricao)",
			actual:  role.Patient,
			allowed: []role.Role{},
		},
		{
			name:    "allowed nil retorna nil (sem restricao)",
			actual:  role.Admin,
			allowed: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := role.MustBeIn(tt.actual, tt.allowed)
			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
