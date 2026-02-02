package searchinput

import (
	"strings"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/list"
)

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

	if err := l.validateSort(); err != nil {
		return err
	}

	if err := l.validateSortConsistency(); err != nil {
		return err
	}

	return nil
}

func (l *ListSearchInput) validateHasSearchCriteria() error {
	hasSearchTerm := l.SearchTerm != nil && strings.TrimSpace(*l.SearchTerm) != ""
	hasFilters := len(l.Filters) > 0

	if !hasSearchTerm && !hasFilters {
		return list.ErrEmptySearchCriteria
	}

	return nil
}

func (l *ListSearchInput) validatePagination() error {
	if l.Pagination == nil {
		return list.ErrMissingPagination
	}
	return nil
}

func (l *ListSearchInput) validateSearchTerm() error {
	if l.SearchTerm == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*l.SearchTerm)
	if trimmed == "" {
		return list.ErrEmptySearchTerm
	}

	const minSearchTermLength = 2
	if len(trimmed) < minSearchTermLength {
		return list.ErrSearchTermTooShort
	}

	const maxSearchTermLength = 100
	if len(trimmed) > maxSearchTermLength {
		return list.ErrSearchTermTooLong
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

func (l *ListSearchInput) HasSort() bool {
	return len(l.Sort) > 0
}

func (l *ListSearchInput) GetFilterByField(field SearchableField) (*Filter, bool) {
	for _, filter := range l.Filters {
		if filter.Field == field {
			return &filter, true
		}
	}
	return nil, false
}

func (l *ListSearchInput) GetPrimarySortField() *SearchableField {
	if len(l.Sort) == 0 {
		return nil
	}
	return &l.Sort[0].Field
}

func (l *ListSearchInput) GetPrimarySortOrder() *SortOrder {
	if len(l.Sort) == 0 {
		return nil
	}
	return &l.Sort[0].Order
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
