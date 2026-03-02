package application

import (
	cursor "github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
)

type SearchSpecialistsDTO struct {
	SearchTerm *string
	Filters    []searchinput.Filter
	Sort       []searchinput.Sort
	Pagination *cursor.CursorPaginationInput
}
