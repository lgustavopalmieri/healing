package searchoutput

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

type SearchResult struct {
	Specialists     []*domain.Specialist
	HasNextPage     bool
	FirstSortValues []any
	LastSortValues  []any
}

type ListSearchOutput struct {
	Specialists  []*domain.Specialist
	CursorOutput *cursor.CursorPaginationOutput
}

func NewListSearchOutput(
	specialists []*domain.Specialist,
	cursorOutput *cursor.CursorPaginationOutput,
) *ListSearchOutput {
	return &ListSearchOutput{
		Specialists:  specialists,
		CursorOutput: cursorOutput,
	}
}
