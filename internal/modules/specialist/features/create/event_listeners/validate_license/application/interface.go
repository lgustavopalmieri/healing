package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

type ValidateLicenseRepositoryInterface interface {
	FindByID(ctx context.Context, id string) (*domain.Specialist, error)
	UpdateStatus(ctx context.Context, id string, status domain.SpecialistStatus) (*domain.Specialist, error)
}

type LicenseGatewayInterface interface {
	Validate(ctx context.Context, licenseNumber string) (bool, error)
}
