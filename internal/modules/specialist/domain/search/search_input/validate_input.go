package searchinput

import (
	"strings"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search"
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
		return search.ErrEmptySearchCriteria
	}

	return nil
}

func (l *ListSearchInput) validatePagination() error {
	if l.Pagination == nil {
		return search.ErrMissingPagination
	}
	return nil
}

func (l *ListSearchInput) validateSearchTerm() error {
	if l.SearchTerm == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*l.SearchTerm)
	if trimmed == "" {
		return search.ErrEmptySearchTerm
	}

	const minSearchTermLength = 2
	if len(trimmed) < minSearchTermLength {
		return search.ErrSearchTermTooShort
	}

	const maxSearchTermLength = 100
	if len(trimmed) > maxSearchTermLength {
		return search.ErrSearchTermTooLong
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

	l.ensureSearchableStatusFilter()
	l.ensureDefaultSort()
}

func (l *ListSearchInput) ensureSearchableStatusFilter() {
	for _, f := range l.Filters {
		if f.Field == FieldStatus {
			return
		}
	}

	l.Filters = append(l.Filters, Filter{
		Field:  FieldStatus,
		Values: domain.SearchableStatuses(),
	})
}

func (l *ListSearchInput) ensureDefaultSort() {
	hasRating := false
	hasUpdatedAt := false

	for _, s := range l.Sort {
		if s.Field == FieldRating {
			hasRating = true
		}
		if s.Field == FieldUpdatedAt {
			hasUpdatedAt = true
		}
	}

	if !hasRating {
		l.Sort = append(l.Sort, Sort{
			Field: FieldRating,
			Order: SortDesc,
		})
	}

	if !hasUpdatedAt {
		l.Sort = append(l.Sort, Sort{
			Field: FieldUpdatedAt,
			Order: SortDesc,
		})
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
