package application

import (
	"context"

	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
)

type SpecialistSearchRepositoryInterface interface {
	Search(ctx context.Context, input *searchinput.ListSearchInput) (*searchoutput.ListSearchOutput, error)
}
