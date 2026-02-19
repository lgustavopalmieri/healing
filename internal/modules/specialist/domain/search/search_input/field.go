package searchinput

import (
	"strings"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search"
)

type SearchableField string

const (
	FieldName        SearchableField = "name"
	FieldSpecialty   SearchableField = "specialty"
	FieldDescription SearchableField = "description"
	FieldKeywords    SearchableField = "keywords"
	FieldRating      SearchableField = "rating"
	FieldCreatedAt   SearchableField = "created_at"
	FieldUpdatedAt   SearchableField = "updated_at"
)

type Filter struct {
	Field SearchableField
	Value string
}

func (l *ListSearchInput) validateFilters() error {
	if len(l.Filters) == 0 {
		return nil
	}

	seenFields := make(map[SearchableField]bool)

	for _, filter := range l.Filters {
		if !filter.Field.IsValid() {
			return search.NewErrInvalidSearchField(string(filter.Field))
		}

		if strings.TrimSpace(filter.Value) == "" {
			return search.NewErrEmptyFilterValue(string(filter.Field))
		}

		if seenFields[filter.Field] {
			return search.NewErrDuplicateFilter(string(filter.Field))
		}
		seenFields[filter.Field] = true
	}

	return nil
}

func (s SearchableField) IsValid() bool {
	switch s {
	case FieldName, FieldSpecialty, FieldDescription,
		FieldKeywords, FieldRating, FieldCreatedAt, FieldUpdatedAt:
		return true
	default:
		return false
	}
}

func (s SearchableField) IsSortable() bool {
	switch s {
	case FieldName, FieldSpecialty, FieldRating, FieldCreatedAt, FieldUpdatedAt:
		return true
	case FieldDescription, FieldKeywords:
		return false
	default:
		return false
	}
}

func (s SearchableField) SupportsCursorPagination() bool {
	switch s {
	case FieldRating, FieldUpdatedAt:
		return true
	case FieldCreatedAt, FieldName, FieldSpecialty:
		return false
	case FieldDescription, FieldKeywords:
		return false
	default:
		return false
	}
}
