package application

import (
	"context"

	cursor "github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
)

func (c *SearchSpecialistsUseCase) Execute(ctx context.Context, dto *SearchSpecialistsDTO) (*searchoutput.ListSearchOutput, error) {
	if dto == nil {
		return nil, ErrInvalidSearchInput
	}

	input, err := searchinput.NewListSearchInput(dto.SearchTerm, dto.Filters, dto.Sort, dto.Pagination)
	if err != nil {
		if search.IsListSearchDomainError(err) {
			return nil, ErrInvalidSearchInput
		}
		return nil, ErrInvalidSearchInput
	}

	result, err := c.repository.Search(ctx, input)
	if err != nil {
		return nil, ErrSearchExecution
	}

	cursorOutput := c.buildPagination(input, result)

	return searchoutput.NewListSearchOutput(result.Specialists, cursorOutput), nil
}

func (c *SearchSpecialistsUseCase) buildPagination(input *searchinput.ListSearchInput, result *searchoutput.SearchResult) *cursor.CursorPaginationOutput {
	var nextCursor *string
	if result.HasNextPage && len(result.LastSortValues) > 0 {
		encoded := cursor.EncodeCursorMultiSort(result.LastSortValues)
		nextCursor = &encoded
	}

	var prevCursor *string
	if !input.Pagination.IsFirstPage() && len(result.FirstSortValues) > 0 {
		encoded := cursor.EncodeCursorMultiSort(result.FirstSortValues)
		prevCursor = &encoded
	}

	return cursor.NewCursorPaginationOutput(
		nextCursor,
		prevCursor,
		result.HasNextPage,
		!input.Pagination.IsFirstPage(),
		len(result.Specialists),
	)
}
