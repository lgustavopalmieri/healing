package searchinput

import "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/list"

type Sort struct {
	Field SearchableField
	Order SortOrder
}

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

func (l *ListSearchInput) validateSort() error {
	if len(l.Sort) == 0 {
		return nil
	}

	seenFields := make(map[SearchableField]bool)

	for _, sort := range l.Sort {
		if !sort.Field.IsValid() {
			return list.NewErrInvalidSearchField(string(sort.Field))
		}

		if !sort.Field.IsSortable() {
			return list.NewErrFieldNotSortable(string(sort.Field))
		}

		if !sort.Order.IsValid() {
			return list.NewErrInvalidSortOrder(string(sort.Order))
		}

		if seenFields[sort.Field] {
			return list.NewErrDuplicateSortCriteria(string(sort.Field))
		}
		seenFields[sort.Field] = true
	}

	return nil
}

func (l *ListSearchInput) validateSortConsistency() error {
	if len(l.Sort) == 0 {
		return nil
	}

	firstSort := l.Sort[0]
	if !firstSort.Field.SupportsCursorPagination() {
		return list.NewErrFieldNotSupportsCursor(string(firstSort.Field))
	}

	return nil
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

func (s SortOrder) IsValid() bool {
	return s == SortAsc || s == SortDesc
}

func (s SortOrder) String() string {
	return string(s)
}
