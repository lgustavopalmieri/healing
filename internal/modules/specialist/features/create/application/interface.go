package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

type SpecialistCreateRepositoryInterface interface {
	Save(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByLicenseNumber(ctx context.Context, licenseNumber string) (bool, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
	ValidateUniqueness(ctx context.Context, id, email, licenseNumber string)
}

type SpecialistCreateExternalGatewayInterface interface {
	ValidateLicenseNumber(ctx context.Context, licenseNumber string) (bool, error)
}
