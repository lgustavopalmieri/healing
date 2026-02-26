package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
)

func (c *SearchSpecialistsCommand) Execute(ctx context.Context, input *searchinput.ListSearchInput) (*searchoutput.ListSearchOutput, error) {
	if input == nil {
		c.logger.Error(ctx, ErrInvalidSearchInputMessage)
		return nil, ErrInvalidSearchInput
	}

	output, err := c.repository.Search(ctx, input)
	if err != nil {
		c.logger.Error(ctx, ErrSearchExecutionMessage,
			observability.Field{Key: "error", Value: err.Error()})

		if search.IsListSearchDomainError(err) {
			return nil, ErrInvalidSearchInput
		}

		return nil, ErrSearchExecution
	}

	return output, nil
}
