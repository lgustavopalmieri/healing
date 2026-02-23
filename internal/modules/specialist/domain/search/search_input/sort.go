package searchinput

import "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search"

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
			return search.NewErrInvalidSearchField(string(sort.Field))
		}

		if !sort.Field.IsSortable() {
			return search.NewErrFieldNotSortable(string(sort.Field))
		}

		if !sort.Order.IsValid() {
			return search.NewErrInvalidSortOrder(string(sort.Order))
		}

		if seenFields[sort.Field] {
			return search.NewErrDuplicateSortCriteria(string(sort.Field))
		}
		seenFields[sort.Field] = true
	}

	return nil
}

func (l *ListSearchInput) validateSortConsistency() error {
	if len(l.Sort) == 0 {
		return nil
	}

	hasCursorCompatibleSort := false
	for _, sort := range l.Sort {
		if sort.Field.SupportsCursorPagination() {
			hasCursorCompatibleSort = true
			break
		}
	}

	if !hasCursorCompatibleSort {
		return search.NewErrFieldNotSupportsCursor(string(l.Sort[0].Field))
	}

	return nil
}

func (s SortOrder) IsValid() bool {
	return s == SortAsc || s == SortDesc
}

func (s SortOrder) String() string {
	return string(s)
}
