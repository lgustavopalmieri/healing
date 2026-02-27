package listener

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

//go:generate mockgen -source=interface.go -destination=mocks/repository_mock.go -package=mocks
type SpecialistFindByIDRepositoryInterface interface {
	FindByID(ctx context.Context, id string) (*domain.Specialist, error)
}

//go:generate mockgen -source=interface.go -destination=mocks/read_projection_mock.go -package=mocks
type SpecialistReadProjectionInterface interface {
	Update(ctx context.Context, specialist *domain.Specialist) error
}
