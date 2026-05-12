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
			name: "happy path - raw valido retorna Password nao-vazia",
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
			_, err := password.NewPassword(tt.raw, tt.cfg)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPassword_HashAndMatches(t *testing.T) {
	t.Run("happy path - hash + match com senha correta retorna true", func(t *testing.T) {
		pwd, err := password.NewPassword("abc12345", defaultCfg())
		require.NoError(t, err)

		hashed, err := pwd.Hash(testBcryptCost)
		require.NoError(t, err)
		require.False(t, hashed.IsEmpty())
		assert.NotEqual(t, "abc12345", hashed.String())

		assert.True(t, pwd.Matches(hashed))
	})

	t.Run("failure - match com senha errada retorna false", func(t *testing.T) {
		pwd, err := password.NewPassword("abc12345", defaultCfg())
		require.NoError(t, err)

		hashed, err := pwd.Hash(testBcryptCost)
		require.NoError(t, err)

		wrongAttempt, err := password.NewPassword("wrong-password-1", defaultCfg())
		require.NoError(t, err)

		assert.False(t, wrongAttempt.Matches(hashed))
	})

	t.Run("failure - match com hash malformado retorna false", func(t *testing.T) {
		pwd, err := password.NewPassword("abc12345", defaultCfg())
		require.NoError(t, err)

		malformed := password.NewHashedPassword("not-a-bcrypt-hash")
		assert.False(t, pwd.Matches(malformed))
	})
}

func TestHashedPassword(t *testing.T) {
	t.Run("empty hashed password", func(t *testing.T) {
		h := password.NewHashedPassword("")
		assert.True(t, h.IsEmpty())
		assert.Equal(t, "", h.String())
	})

	t.Run("non-empty hashed password", func(t *testing.T) {
		h := password.NewHashedPassword("hash-123")
		assert.False(t, h.IsEmpty())
		assert.Equal(t, "hash-123", h.String())
	})
}
