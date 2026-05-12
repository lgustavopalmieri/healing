package credential

import (
	"time"

	"github.com/google/uuid"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/password"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

type Credential struct {
	ID             string
	SubjectID      string
	Role           role.Role
	Provider       provider.Provider
	ProviderUserID string
	PasswordHash   password.HashedPassword
	Email          string
	Status         Status
	LastUsedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type NewCredentialInput struct {
	SubjectID      string
	Role           role.Role
	Provider       provider.Provider
	ProviderUserID string
	Email          string
}

func NewCredential(in NewCredentialInput) *Credential {
	now := time.Now().UTC()
	return &Credential{
		ID:             uuid.New().String(),
		SubjectID:      in.SubjectID,
		Role:           in.Role,
		Provider:       in.Provider,
		ProviderUserID: in.ProviderUserID,
		Email:          in.Email,
		Status:         StatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (c *Credential) Activate(passwordHash password.HashedPassword) error {
	if !c.Status.CanTransitionTo(StatusActive) {
		return ErrInvalidStatusTransition
	}
	c.PasswordHash = passwordHash
	c.Status = StatusActive
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Credential) UpdatePassword(passwordHash password.HashedPassword) error {
	if c.Status != StatusActive {
		return ErrInvalidStatusTransition
	}
	c.PasswordHash = passwordHash
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Credential) Lock() error {
	if !c.Status.CanTransitionTo(StatusLocked) {
		return ErrInvalidStatusTransition
	}
	c.Status = StatusLocked
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Credential) MarkUsed() {
	now := time.Now().UTC()
	c.LastUsedAt = &now
}
