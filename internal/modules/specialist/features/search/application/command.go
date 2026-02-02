package application

import (
	"context"
	"fmt"

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
	} else {
		c.logger.Info(ctx, SearchCompletedMessage,
			observability.Field{Key: "resultsCount", Value: intToString(output.CursorOutput.TotalItemsInPage)},
			observability.Field{Key: "hasNextPage", Value: boolToString(output.CursorOutput.HasNextPage)})
	}

	return output, nil
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}
