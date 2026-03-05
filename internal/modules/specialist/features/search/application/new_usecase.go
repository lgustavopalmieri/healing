package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type SearchSpecialistsUseCase struct {
	repository SpecialistSearchRepositoryInterface
	logger     observability.Logger
}

func NewSearchSpecialistsUseCase(
	repository SpecialistSearchRepositoryInterface,
	logger observability.Logger,
) *SearchSpecialistsUseCase {
	return &SearchSpecialistsUseCase{
		repository: repository,
		logger:     logger,
	}
}
