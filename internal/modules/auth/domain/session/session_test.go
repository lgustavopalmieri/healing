package session_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func sessionFactory(overrides ...func(*session.NewSessionInput)) session.NewSessionInput {
	in := session.NewSessionInput{
		SubjectID:        "subject-1",
		Role:             role.Specialist,
		RefreshTokenHash: "hash-abc",
		DeviceInfo:       "web",
		IPAddress:        "1.2.3.4",
		UserAgent:        "go-test",
		ExpiresAt:        time.Now().UTC().Add(24 * time.Hour),
	}
	for _, o := range overrides {
		o(&in)
	}
	return in
}

func TestSession_New(t *testing.T) {
	t.Run("happy path - cria session com ID UUID, CreatedAt preenchido e RevokedAt nil", func(t *testing.T) {
		in := sessionFactory()
		s := session.NewSession(in)

		require.NotNil(t, s)

		_, err := uuid.Parse(s.ID)
		assert.NoError(t, err)

		assert.Equal(t, in.SubjectID, s.SubjectID)
		assert.Equal(t, in.Role, s.Role)
		assert.Equal(t, in.RefreshTokenHash, s.RefreshTokenHash)
		assert.Equal(t, in.DeviceInfo, s.DeviceInfo)
		assert.Equal(t, in.IPAddress, s.IPAddress)
		assert.Equal(t, in.UserAgent, s.UserAgent)
		assert.Equal(t, in.ExpiresAt, s.ExpiresAt)
		assert.Nil(t, s.RevokedAt)
		assert.Nil(t, s.LastUsedAt)
		assert.False(t, s.CreatedAt.IsZero())
	})
}

func TestSession_Revoke(t *testing.T) {
	t.Run("happy path - primeira chamada seta RevokedAt", func(t *testing.T) {
		s := session.NewSession(sessionFactory())
		assert.Nil(t, s.RevokedAt)

		err := s.Revoke()

		require.NoError(t, err)
		require.NotNil(t, s.RevokedAt)
		assert.WithinDuration(t, time.Now().UTC(), *s.RevokedAt, 1*time.Second)
	})

	t.Run("failure - segunda chamada retorna ErrAlreadyRevoked", func(t *testing.T) {
		s := session.NewSession(sessionFactory())
		require.NoError(t, s.Revoke())

		err := s.Revoke()

		require.Error(t, err)
		assert.ErrorIs(t, err, session.ErrAlreadyRevoked)
	})
}

func TestSession_IsRevoked(t *testing.T) {
	t.Run("false quando RevokedAt nil", func(t *testing.T) {
		s := session.NewSession(sessionFactory())
		assert.False(t, s.IsRevoked())
	})

	t.Run("true apos Revoke()", func(t *testing.T) {
		s := session.NewSession(sessionFactory())
		require.NoError(t, s.Revoke())
		assert.True(t, s.IsRevoked())
	})
}

func TestSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		now       time.Time
		expected  bool
	}{
		{
			name:      "false quando now < ExpiresAt",
			expiresAt: time.Now().Add(1 * time.Hour),
			now:       time.Now(),
			expected:  false,
		},
		{
			name:      "true quando now > ExpiresAt",
			expiresAt: time.Now().Add(-1 * time.Hour),
			now:       time.Now(),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := session.NewSession(sessionFactory(func(in *session.NewSessionInput) {
				in.ExpiresAt = tt.expiresAt
			}))
			assert.Equal(t, tt.expected, s.IsExpired(tt.now))
		})
	}
}

func TestSession_MarkUsed(t *testing.T) {
	t.Run("happy path - seta LastUsedAt", func(t *testing.T) {
		s := session.NewSession(sessionFactory())
		assert.Nil(t, s.LastUsedAt)

		s.MarkUsed()

		require.NotNil(t, s.LastUsedAt)
		assert.WithinDuration(t, time.Now().UTC(), *s.LastUsedAt, 1*time.Second)
	})
}
