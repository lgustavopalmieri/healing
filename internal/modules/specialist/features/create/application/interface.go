package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

type SpecialistCreateRepositoryInterface interface {
	SaveWithValidation(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error)
}
