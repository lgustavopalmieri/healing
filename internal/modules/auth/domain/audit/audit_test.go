package audit_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/audit"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func TestAuditLog_New(t *testing.T) {
	t.Run("happy path - cria log com ID UUID e campos ecoados", func(t *testing.T) {
		in := audit.NewAuditLogInput{
			SubjectID: "subject-1",
			Role:      role.Specialist,
			EventType: audit.EventLoginSuccess,
			IPAddress: "1.2.3.4",
			UserAgent: "go-test",
			Metadata: map[string]any{
				"provider": "password",
			},
		}

		log := audit.NewAuditLog(in)

		require.NotNil(t, log)
		_, err := uuid.Parse(log.ID)
		assert.NoError(t, err)

		assert.Equal(t, in.SubjectID, log.SubjectID)
		assert.Equal(t, in.Role, log.Role)
		assert.Equal(t, in.EventType, log.EventType)
		assert.Equal(t, in.IPAddress, log.IPAddress)
		assert.Equal(t, in.UserAgent, log.UserAgent)
		assert.Equal(t, in.Metadata, log.Metadata)
		assert.WithinDuration(t, time.Now().UTC(), log.CreatedAt, 1*time.Second)
	})

	t.Run("happy path - preserva metadata nil como nil", func(t *testing.T) {
		log := audit.NewAuditLog(audit.NewAuditLogInput{
			SubjectID: "",
			EventType: audit.EventAccessDenied,
			IPAddress: "1.2.3.4",
		})

		require.NotNil(t, log)
		assert.Nil(t, log.Metadata)
	})
}

func TestEventType_Category(t *testing.T) {
	tests := []struct {
		event    audit.EventType
		expected audit.Category
	}{
		{event: audit.EventLoginSuccess, expected: audit.CategoryAuthentication},
		{event: audit.EventLoginFailure, expected: audit.CategoryAuthentication},
		{event: audit.EventLogout, expected: audit.CategoryAuthentication},

		{event: audit.EventPasswordSet, expected: audit.CategoryPassword},
		{event: audit.EventPasswordReset, expected: audit.CategoryPassword},
		{event: audit.EventPasswordChanged, expected: audit.CategoryPassword},
		{event: audit.EventPasswordResetRequested, expected: audit.CategoryPassword},

		{event: audit.EventAdminAccessResource, expected: audit.CategoryAdmin},

		{event: audit.EventSessionRevoked, expected: audit.CategorySecurity},
		{event: audit.EventRevokeAllSessions, expected: audit.CategorySecurity},
		{event: audit.EventCredentialLocked, expected: audit.CategorySecurity},
		{event: audit.EventAccessDenied, expected: audit.CategorySecurity},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.event.Category())
		})
	}
}
