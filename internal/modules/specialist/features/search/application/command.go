package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
)

func (c *SearchSpecialistsCommand) Execute(ctx context.Context, input *searchinput.ListSearchInput) (*searchoutput.ListSearchOutput, error) {
	output, err := c.repository.Search(ctx, input)
	if err != nil {
		c.logger.Error(ctx, ErrSearchExecutionMessage,
			observability.Field{Key: "error", Value: err.Error()})
		return nil, ErrSearchExecution
	}

	if output.CursorOutput.IsEmpty() {
		c.logger.Info(ctx, SearchNoResultsMessage)
	}

	return output, nil
}
