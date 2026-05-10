package credential_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func credFactory(overrides ...func(*credential.NewCredentialInput)) credential.NewCredentialInput {
	in := credential.NewCredentialInput{
		SubjectID: "subject-1",
		Role:      role.Specialist,
		Provider:  provider.Password,
		Email:     "user@healing.com",
	}
	for _, o := range overrides {
		o(&in)
	}
	return in
}

func TestCredential_New(t *testing.T) {
	t.Run("happy path - cria credential pending com ID UUID e timestamps", func(t *testing.T) {
		in := credFactory()
		cred := credential.NewCredential(in)

		require.NotNil(t, cred)

		_, err := uuid.Parse(cred.ID)
		assert.NoError(t, err, "ID should be a valid UUID")

		assert.Equal(t, in.SubjectID, cred.SubjectID)
		assert.Equal(t, in.Role, cred.Role)
		assert.Equal(t, in.Provider, cred.Provider)
		assert.Equal(t, in.Email, cred.Email)
		assert.Equal(t, credential.StatusPending, cred.Status)
		assert.Empty(t, cred.PasswordHash)
		assert.Nil(t, cred.LastUsedAt)
		assert.False(t, cred.CreatedAt.IsZero())
		assert.False(t, cred.UpdatedAt.IsZero())
	})
}

func TestCredential_Activate(t *testing.T) {
	tests := []struct {
		name         string
		initial      credential.Status
		expectError  bool
		expectedErr  error
		expectedNext credential.Status
	}{
		{
			name:         "happy path - pending -> active",
			initial:      credential.StatusPending,
			expectedNext: credential.StatusActive,
		},
		{
			name:         "happy path - locked -> active (admin destrava conta)",
			initial:      credential.StatusLocked,
			expectedNext: credential.StatusActive,
		},
		{
			name:        "failure - active -> active retorna ErrInvalidStatusTransition",
			initial:     credential.StatusActive,
			expectError: true,
			expectedErr: credential.ErrInvalidStatusTransition,
		},
		{
			name:        "failure - deleted -> active retorna ErrInvalidStatusTransition",
			initial:     credential.StatusDeleted,
			expectError: true,
			expectedErr: credential.ErrInvalidStatusTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := credential.NewCredential(credFactory())
			cred.Status = tt.initial
			originalUpdatedAt := cred.UpdatedAt

			time.Sleep(time.Millisecond)
			err := cred.Activate("hash-xyz")

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Equal(t, tt.initial, cred.Status)
				assert.Empty(t, cred.PasswordHash)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedNext, cred.Status)
			assert.Equal(t, "hash-xyz", cred.PasswordHash)
			assert.True(t, cred.UpdatedAt.After(originalUpdatedAt))
		})
	}
}

func TestCredential_UpdatePassword(t *testing.T) {
	tests := []struct {
		name        string
		initial     credential.Status
		expectError bool
	}{
		{name: "happy path - active atualiza hash", initial: credential.StatusActive},
		{name: "failure - pending retorna ErrInvalidStatusTransition", initial: credential.StatusPending, expectError: true},
		{name: "failure - locked retorna ErrInvalidStatusTransition", initial: credential.StatusLocked, expectError: true},
		{name: "failure - deleted retorna ErrInvalidStatusTransition", initial: credential.StatusDeleted, expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := credential.NewCredential(credFactory())
			cred.Status = tt.initial
			cred.PasswordHash = "old-hash"
			originalUpdatedAt := cred.UpdatedAt

			time.Sleep(time.Millisecond)
			err := cred.UpdatePassword("new-hash")

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, credential.ErrInvalidStatusTransition)
				assert.Equal(t, "old-hash", cred.PasswordHash)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "new-hash", cred.PasswordHash)
			assert.True(t, cred.UpdatedAt.After(originalUpdatedAt))
		})
	}
}

func TestCredential_Lock(t *testing.T) {
	tests := []struct {
		name        string
		initial     credential.Status
		expectError bool
	}{
		{name: "happy path - active -> locked", initial: credential.StatusActive},
		{name: "failure - pending -> locked retorna ErrInvalidStatusTransition", initial: credential.StatusPending, expectError: true},
		{name: "failure - locked -> locked retorna ErrInvalidStatusTransition", initial: credential.StatusLocked, expectError: true},
		{name: "failure - deleted -> locked retorna ErrInvalidStatusTransition", initial: credential.StatusDeleted, expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := credential.NewCredential(credFactory())
			cred.Status = tt.initial

			err := cred.Lock()

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, credential.ErrInvalidStatusTransition)
				assert.Equal(t, tt.initial, cred.Status)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, credential.StatusLocked, cred.Status)
		})
	}
}

func TestCredential_MarkUsed(t *testing.T) {
	t.Run("happy path - seta LastUsedAt para agora", func(t *testing.T) {
		cred := credential.NewCredential(credFactory())
		assert.Nil(t, cred.LastUsedAt)

		cred.MarkUsed()

		require.NotNil(t, cred.LastUsedAt)
		assert.WithinDuration(t, time.Now().UTC(), *cred.LastUsedAt, 1*time.Second)
	})
}
