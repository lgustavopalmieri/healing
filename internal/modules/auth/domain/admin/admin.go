package admin

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	ID        string
	Name      string
	Email     string
	SubRole   SubRole
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NewAdminInput struct {
	Name    string
	Email   string
	SubRole SubRole
}

func NewAdmin(in NewAdminInput) (*Admin, error) {
	if !in.SubRole.Valid() {
		return nil, ErrInvalidSubRole
	}
	now := time.Now().UTC()
	return &Admin{
		ID:        uuid.New().String(),
		Name:      in.Name,
		Email:     in.Email,
		SubRole:   in.SubRole,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (a *Admin) Activate() error {
	if !a.Status.CanTransitionTo(StatusActive) {
		return ErrInvalidStatusTransition
	}
	a.Status = StatusActive
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func (a *Admin) Deactivate() error {
	if !a.Status.CanTransitionTo(StatusInactive) {
		return ErrInvalidStatusTransition
	}
	a.Status = StatusInactive
	a.UpdatedAt = time.Now().UTC()
	return nil
}
