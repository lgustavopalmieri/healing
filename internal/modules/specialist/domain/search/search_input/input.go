package searchinput

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
)

type ListSearchInput struct {
	SearchTerm *string
	Filters    []Filter
	Sort       []Sort
	Pagination *cursor.CursorPaginationInput
}

func NewListSearchInput(
	searchTerm *string,
	filters []Filter,
	sort []Sort,
	pagination *cursor.CursorPaginationInput,
) (*ListSearchInput, error) {
	input := &ListSearchInput{
		SearchTerm: searchTerm,
		Filters:    filters,
		Sort:       sort,
		Pagination: pagination,
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	input.normalize()

	return input, nil
}
