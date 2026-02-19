package application

import (
	"context"

	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
)

//go:generate mockgen -source=interface.go -destination=mocks/repository_mock.go -package=mocks
type SpecialistSearchRepositoryInterface interface {
	Search(ctx context.Context, input *searchinput.ListSearchInput) (*searchoutput.ListSearchOutput, error)
}
