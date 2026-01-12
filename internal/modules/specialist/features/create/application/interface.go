package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

type SpecialistCreateRepositoryInterface interface {
	Save(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error)
	ValidateUniqueness(ctx context.Context, id, email, licenseNumber string) error
}

type SpecialistCreateExternalGatewayInterface interface {
	ValidateLicenseNumber(ctx context.Context, licenseNumber string) (bool, error)
}
