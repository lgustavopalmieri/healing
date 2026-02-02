package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type SearchSpecialistsCommand struct {
	repository SpecialistSearchRepositoryInterface
	logger     observability.Logger
}

func NewSearchSpecialistsCommand(
	repository SpecialistSearchRepositoryInterface,
	logger observability.Logger,
) *SearchSpecialistsCommand {
	return &SearchSpecialistsCommand{
		repository: repository,
		logger:     logger,
	}
}
