package audit

import (
	"time"

	"github.com/google/uuid"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type AuditLog struct {
	ID        string
	SubjectID string
	Role      role.Role
	EventType EventType
	IPAddress string
	UserAgent string
	Metadata  map[string]any
	CreatedAt time.Time
}

type NewAuditLogInput struct {
	SubjectID string
	Role      role.Role
	EventType EventType
	IPAddress string
	UserAgent string
	Metadata  map[string]any
}

func NewAuditLog(in NewAuditLogInput) *AuditLog {
	return &AuditLog{
		ID:        uuid.New().String(),
		SubjectID: in.SubjectID,
		Role:      in.Role,
		EventType: in.EventType,
		IPAddress: in.IPAddress,
		UserAgent: in.UserAgent,
		Metadata:  in.Metadata,
		CreatedAt: time.Now().UTC(),
	}
}
