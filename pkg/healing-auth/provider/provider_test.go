package provider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
)

func TestParse_Provider(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    provider.Provider
		expectError bool
		expectedErr error
	}{
		{
			name:     "happy path - password retorna Provider password",
			input:    "password",
			expected: provider.Password,
		},
		{
			name:     "happy path - google retorna Provider google",
			input:    "google",
			expected: provider.Google,
		},
		{
			name:     "happy path - instagram retorna Provider instagram",
			input:    "instagram",
			expected: provider.Instagram,
		},
		{
			name:     "happy path - biometric retorna Provider biometric",
			input:    "biometric",
			expected: provider.Biometric,
		},
		{
			name:        "failure - string invalida retorna ErrInvalidProvider",
			input:       "facebook",
			expectError: true,
			expectedErr: provider.ErrInvalidProvider,
		},
		{
			name:        "failure - string vazia retorna ErrInvalidProvider",
			input:       "",
			expectError: true,
			expectedErr: provider.ErrInvalidProvider,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Parse(tt.input)
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

func TestProvider_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    provider.Provider
		expected bool
	}{
		{name: "true para password", input: provider.Password, expected: true},
		{name: "true para google", input: provider.Google, expected: true},
		{name: "true para instagram", input: provider.Instagram, expected: true},
		{name: "true para biometric", input: provider.Biometric, expected: true},
		{name: "false para string desconhecida", input: provider.Provider("apple"), expected: false},
		{name: "false para string vazia", input: provider.Provider(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.Valid())
		})
	}
}

func TestProvider_Methods(t *testing.T) {
	tests := []struct {
		name             string
		input            provider.Provider
		requiresPassword bool
		isExternal       bool
	}{
		{
			name:             "RequiresPassword true apenas para password",
			input:            provider.Password,
			requiresPassword: true,
			isExternal:       false,
		},
		{
			name:             "IsExternal true para google",
			input:            provider.Google,
			requiresPassword: false,
			isExternal:       true,
		},
		{
			name:             "IsExternal true para instagram",
			input:            provider.Instagram,
			requiresPassword: false,
			isExternal:       true,
		},
		{
			name:             "biometric nao requer password e nao e externo",
			input:            provider.Biometric,
			requiresPassword: false,
			isExternal:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.requiresPassword, tt.input.RequiresPassword())
			assert.Equal(t, tt.isExternal, tt.input.IsExternal())
		})
	}
}

func TestProvider_String(t *testing.T) {
	assert.Equal(t, "password", provider.Password.String())
	assert.Equal(t, "google", provider.Google.String())
}
