package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

//go:generate mockgen -source=interface.go -destination=mocks/repository_mock.go -package=mocks
type SpecialistUpdateRepositoryInterface interface {
	FindByID(ctx context.Context, id string) (*domain.Specialist, error)
	Update(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error)
}
