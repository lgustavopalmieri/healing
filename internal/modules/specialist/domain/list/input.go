package list

import (
	"strings"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
)

type ListSearchInput struct {
	SearchTerm *string
	Filters    []Filter
	Sorting    []SortCriteria
	Pagination *cursor.CursorPaginationInput
}

type Filter struct {
	Field SearchableField
	Value string
}

type SortCriteria struct {
	Field SearchableField
	Order SortOrder
}

type SearchableField string

const (
	FieldName        SearchableField = "name"
	FieldSpecialty   SearchableField = "specialty"
	FieldDescription SearchableField = "description"
	FieldKeywords    SearchableField = "keywords"
	FieldCreatedAt   SearchableField = "created_at"
	FieldUpdatedAt   SearchableField = "updated_at"
)

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

func NewListSearchInput(
	searchTerm *string,
	filters []Filter,
	sorting []SortCriteria,
	pagination *cursor.CursorPaginationInput,
) (*ListSearchInput, error) {
	input := &ListSearchInput{
		SearchTerm: searchTerm,
		Filters:    filters,
		Sorting:    sorting,
		Pagination: pagination,
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	input.normalize()

	return input, nil
}

func (l *ListSearchInput) validate() error {
	if err := l.validateHasSearchCriteria(); err != nil {
		return err
	}

	if err := l.validatePagination(); err != nil {
		return err
	}

	if err := l.validateSearchTerm(); err != nil {
		return err
	}

	if err := l.validateFilters(); err != nil {
		return err
	}

	if err := l.validateSorting(); err != nil {
		return err
	}

	if err := l.validateSortingConsistency(); err != nil {
		return err
	}

	return nil
}

func (l *ListSearchInput) validateHasSearchCriteria() error {
	hasSearchTerm := l.SearchTerm != nil && strings.TrimSpace(*l.SearchTerm) != ""
	hasFilters := len(l.Filters) > 0

	if !hasSearchTerm && !hasFilters {
		return ErrEmptySearchCriteria
	}

	return nil
}

func (l *ListSearchInput) validatePagination() error {
	if l.Pagination == nil {
		return ErrMissingPagination
	}
	return nil
}

func (l *ListSearchInput) validateSearchTerm() error {
	if l.SearchTerm == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*l.SearchTerm)
	if trimmed == "" {
		return ErrEmptySearchTerm
	}

	const minSearchTermLength = 2
	if len(trimmed) < minSearchTermLength {
		return ErrSearchTermTooShort
	}

	const maxSearchTermLength = 100
	if len(trimmed) > maxSearchTermLength {
		return ErrSearchTermTooLong
	}

	return nil
}

func (l *ListSearchInput) validateFilters() error {
	if len(l.Filters) == 0 {
		return nil
	}

	seenFields := make(map[SearchableField]bool)

	for _, filter := range l.Filters {
		if !filter.Field.IsValid() {
			return NewErrInvalidSearchField(string(filter.Field))
		}

		if !filter.Field.IsFilterable() {
			return NewErrFieldNotFilterable(string(filter.Field))
		}

		if strings.TrimSpace(filter.Value) == "" {
			return NewErrEmptyFilterValue(string(filter.Field))
		}

		if seenFields[filter.Field] {
			return NewErrDuplicateFilter(string(filter.Field))
		}
		seenFields[filter.Field] = true
	}

	return nil
}

func (l *ListSearchInput) validateSorting() error {
	if len(l.Sorting) == 0 {
		return nil
	}

	seenFields := make(map[SearchableField]bool)

	for _, sort := range l.Sorting {
		if !sort.Field.IsValid() {
			return NewErrInvalidSearchField(string(sort.Field))
		}

		if !sort.Field.IsSortable() {
			return NewErrFieldNotSortable(string(sort.Field))
		}

		if !sort.Order.IsValid() {
			return NewErrInvalidSortOrder(string(sort.Order))
		}

		if seenFields[sort.Field] {
			return NewErrDuplicateSortCriteria(string(sort.Field))
		}
		seenFields[sort.Field] = true
	}

	return nil
}

func (l *ListSearchInput) validateSortingConsistency() error {
	if len(l.Sorting) == 0 {
		return nil
	}

	firstSort := l.Sorting[0]
	if !firstSort.Field.SupportsCursorPagination() {
		return NewErrFieldNotSupportsCursor(string(firstSort.Field))
	}

	return nil
}

func (l *ListSearchInput) normalize() {
	if l.SearchTerm != nil {
		trimmed := strings.TrimSpace(*l.SearchTerm)
		l.SearchTerm = &trimmed
	}

	for i := range l.Filters {
		l.Filters[i].Value = strings.TrimSpace(l.Filters[i].Value)
	}
}

func (l *ListSearchInput) HasSearchTerm() bool {
	return l.SearchTerm != nil && *l.SearchTerm != ""
}

func (l *ListSearchInput) HasFilters() bool {
	return len(l.Filters) > 0
}

func (l *ListSearchInput) HasSorting() bool {
	return len(l.Sorting) > 0
}

// func (l *ListSearchInput) GetFilterByField(field SearchableField) (*Filter, bool) {
// 	for _, filter := range l.Filters {
// 		if filter.Field == field {
// 			return &filter, true
// 		}
// 	}
// 	return nil, false
// }

func (l *ListSearchInput) GetPrimarySortField() *SearchableField {
	if len(l.Sorting) == 0 {
		return nil
	}
	return &l.Sorting[0].Field
}

func (l *ListSearchInput) GetPrimarySortOrder() *SortOrder {
	if len(l.Sorting) == 0 {
		return nil
	}
	return &l.Sorting[0].Order
}

func (s SearchableField) IsValid() bool {
	switch s {
	case FieldName, FieldSpecialty, FieldDescription,
		FieldKeywords, FieldCreatedAt, FieldUpdatedAt:
		return true
	default:
		return false
	}
}

func (s SearchableField) IsFilterable() bool {
	switch s {
	case FieldName, FieldSpecialty:
		return true
	case FieldDescription, FieldKeywords:
		return false
	case FieldCreatedAt, FieldUpdatedAt:
		return false
	default:
		return false
	}
}

func (s SearchableField) IsSortable() bool {
	switch s {
	case FieldName, FieldSpecialty, FieldCreatedAt, FieldUpdatedAt:
		return true
	case FieldDescription, FieldKeywords:
		return false
	default:
		return false
	}
}

func (s SearchableField) SupportsCursorPagination() bool {
	switch s {
	case FieldCreatedAt, FieldUpdatedAt:
		return true
	case FieldName, FieldSpecialty:
		return false
	case FieldDescription, FieldKeywords:
		return false
	default:
		return false
	}
}

func (s SearchableField) String() string {
	return string(s)
}

func (s SortOrder) IsValid() bool {
	return s == SortAsc || s == SortDesc
}

func (s SortOrder) String() string {
	return string(s)
}
