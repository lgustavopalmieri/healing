package password_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/password"
)

const testBcryptCost = 4

func defaultCfg() password.ValidationConfig {
	return password.ValidationConfig{MinLength: 8}
}

func TestPassword_New(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		cfg         password.ValidationConfig
		expectError bool
		expectedErr error
	}{
		{
			name: "happy path - raw valido retorna Password com Raw preenchido",
			raw:  "abc12345",
			cfg:  defaultCfg(),
		},
		{
			name: "happy path - password com letras digitos e simbolos passa",
			raw:  "abc@12345",
			cfg:  defaultCfg(),
		},
		{
			name:        "failure - raw curto retorna ErrTooShort",
			raw:         "ab12",
			cfg:         defaultCfg(),
			expectError: true,
			expectedErr: password.ErrTooShort,
		},
		{
			name:        "failure - password vazia retorna ErrTooShort",
			raw:         "",
			cfg:         defaultCfg(),
			expectError: true,
			expectedErr: password.ErrTooShort,
		},
		{
			name:        "failure - password so com letras retorna ErrMissingRequiredChars",
			raw:         "abcdefgh",
			cfg:         defaultCfg(),
			expectError: true,
			expectedErr: password.ErrMissingRequiredChars,
		},
		{
			name:        "failure - password so com digitos retorna ErrMissingRequiredChars",
			raw:         "12345678",
			cfg:         defaultCfg(),
			expectError: true,
			expectedErr: password.ErrMissingRequiredChars,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := password.NewPassword(tt.raw, tt.cfg)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Empty(t, got.Raw)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.raw, got.Raw)
		})
	}
}

func TestPassword_HashAndMatches(t *testing.T) {
	t.Run("happy path - hash + match com senha correta retorna true", func(t *testing.T) {
		pwd, err := password.NewPassword("abc12345", defaultCfg())
		require.NoError(t, err)

		hash, err := pwd.Hash(testBcryptCost)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
		assert.NotEqual(t, "abc12345", hash)

		assert.True(t, password.Matches("abc12345", hash))
	})

	t.Run("failure - match com senha errada retorna false", func(t *testing.T) {
		pwd, err := password.NewPassword("abc12345", defaultCfg())
		require.NoError(t, err)

		hash, err := pwd.Hash(testBcryptCost)
		require.NoError(t, err)

		assert.False(t, password.Matches("wrong-password", hash))
	})

	t.Run("failure - match com hash malformado retorna false", func(t *testing.T) {
		assert.False(t, password.Matches("abc12345", "not-a-bcrypt-hash"))
	})
}
