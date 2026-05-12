package session

import (
	"time"

	"github.com/google/uuid"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type Session struct {
	ID               string
	SubjectID        string
	Role             role.Role
	RefreshTokenHash string
	DeviceInfo       string
	IPAddress        string
	UserAgent        string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	LastUsedAt       *time.Time
	CreatedAt        time.Time
}

type NewSessionInput struct {
	SubjectID        string
	Role             role.Role
	RefreshTokenHash string
	DeviceInfo       string
	IPAddress        string
	UserAgent        string
	ExpiresAt        time.Time
}

func NewSession(in NewSessionInput) *Session {
	return &Session{
		ID:               uuid.New().String(),
		SubjectID:        in.SubjectID,
		Role:             in.Role,
		RefreshTokenHash: in.RefreshTokenHash,
		DeviceInfo:       in.DeviceInfo,
		IPAddress:        in.IPAddress,
		UserAgent:        in.UserAgent,
		ExpiresAt:        in.ExpiresAt,
		CreatedAt:        time.Now().UTC(),
	}
}

func (s *Session) Revoke() error {
	if s.RevokedAt != nil {
		return ErrAlreadyRevoked
	}
	now := time.Now().UTC()
	s.RevokedAt = &now
	return nil
}

func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *Session) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

func (s *Session) MarkUsed() {
	now := time.Now().UTC()
	s.LastUsedAt = &now
}
