package searchinput

import (
	"strings"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/list"
)

type SearchableField string

const (
	FieldName        SearchableField = "name"
	FieldSpecialty   SearchableField = "specialty"
	FieldDescription SearchableField = "description"
	FieldKeywords    SearchableField = "keywords"
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
			return list.NewErrInvalidSearchField(string(filter.Field))
		}

		if !filter.Field.IsFilterable() {
			return list.NewErrFieldNotFilterable(string(filter.Field))
		}

		if strings.TrimSpace(filter.Value) == "" {
			return list.NewErrEmptyFilterValue(string(filter.Field))
		}

		if seenFields[filter.Field] {
			return list.NewErrDuplicateFilter(string(filter.Field))
		}
		seenFields[filter.Field] = true
	}

	return nil
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
