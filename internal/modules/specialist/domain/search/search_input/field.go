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
	FieldStatus      SearchableField = "status"
	FieldCreatedAt   SearchableField = "created_at"
	FieldUpdatedAt   SearchableField = "updated_at"
)

type Filter struct {
	Field  SearchableField
	Value  string
	Values []string
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

		hasValue := strings.TrimSpace(filter.Value) != ""
		hasValues := len(filter.Values) > 0

		if !hasValue && !hasValues {
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
		FieldKeywords, FieldRating, FieldStatus, FieldCreatedAt, FieldUpdatedAt:
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
	case FieldRating, FieldUpdatedAt, FieldCreatedAt:
		return true
	case FieldName, FieldSpecialty:
		return false
	case FieldDescription, FieldKeywords:
		return false
	default:
		return false
	}
}
