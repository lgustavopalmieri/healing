package list

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/offset"
)

// ListSearchInput represents search parameters
type ListSearchInput struct {
	SearchTerm *string
	Filters    []Filter
	Sorting    []SortCriteria
	Pagination cursor.CursorPaginationInput
}

// Filter represents a single filter criterion
type Filter struct {
	Field SearchableField
	Value string
}

// SortCriteria defines sorting behavior
type SortCriteria struct {
	Field SearchableField
	Order SortOrder
}

// SearchableField defines fields that can be searched/filtered
type SearchableField string

const (
	FieldName        SearchableField = "name"
	FieldEmail       SearchableField = "email"
	FieldSpecialty   SearchableField = "specialty"
	FieldDescription SearchableField = "description"
	FieldKeywords    SearchableField = "keywords"
)

// SortOrder defines sort direction
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

