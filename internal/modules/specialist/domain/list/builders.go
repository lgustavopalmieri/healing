package list

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
)

// ListSearchInputBuilder é um builder para construir ListSearchInput
// de forma fluente e legível.
//
// O builder facilita a construção de inputs complexos, especialmente
// quando há múltiplos filtros e critérios de ordenação.
//
// Exemplo de uso:
//
//	searchTerm := "cardiologia"
//	input, err := list.NewListSearchInputBuilder().
//	    WithSearchTerm(&searchTerm).
//	    WithFilter(list.FieldSpecialty, "Cardiology").
//	    WithFilter(list.FieldName, "Dr. Silva").
//	    WithSorting(list.FieldCreatedAt, list.SortDesc).
//	    WithSorting(list.FieldName, list.SortAsc).
//	    WithPagination(paginationInput).
//	    Build()
//
//	if err != nil {
//	    // Tratar erro de validação
//	}
type ListSearchInputBuilder struct {
	SearchTerm *string
	Filters    []Filter
	Sorting    []SortCriteria
	Pagination *cursor.CursorPaginationInput
}

// NewListSearchInputBuilder cria um novo builder.
func NewListSearchInputBuilder() *ListSearchInputBuilder {
	return &ListSearchInputBuilder{
		Filters: make([]Filter, 0),
		Sorting: make([]SortCriteria, 0),
	}
}

func (b *ListSearchInputBuilder) WithSearchTerm(searchTerm *string) *ListSearchInputBuilder {
	b.SearchTerm = searchTerm
	return b
}

func (b *ListSearchInputBuilder) WithFilter(field SearchableField, value string) *ListSearchInputBuilder {
	b.Filters = append(b.Filters, Filter{
		Field: field,
		Value: value,
	})
	return b
}

func (b *ListSearchInputBuilder) WithFilters(filters []Filter) *ListSearchInputBuilder {
	b.Filters = append(b.Filters, filters...)
	return b
}

func (b *ListSearchInputBuilder) WithSorting(field SearchableField, order SortOrder) *ListSearchInputBuilder {
	b.Sorting = append(b.Sorting, SortCriteria{
		Field: field,
		Order: order,
	})
	return b
}

func (b *ListSearchInputBuilder) WithSortingCriteria(sorting []SortCriteria) *ListSearchInputBuilder {
	b.Sorting = append(b.Sorting, sorting...)
	return b
}

func (b *ListSearchInputBuilder) WithPagination(pagination *cursor.CursorPaginationInput) *ListSearchInputBuilder {
	b.Pagination = pagination
	return b
}

func (b *ListSearchInputBuilder) Build() (*ListSearchInput, error) {
	return NewListSearchInput(
		b.SearchTerm,
		b.Filters,
		b.Sorting,
		b.Pagination,
	)
}

// FilterBuilder é um builder para construir filtros de forma fluente.
//
// Exemplo de uso:
//
//	filter := list.NewFilterBuilder().
//	    ForField(list.FieldSpecialty).
//	    WithValue("Cardiology").
//	    Build()
type FilterBuilder struct {
	Field SearchableField
	Value string
}

func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{}
}

func (b *FilterBuilder) ForField(field SearchableField) *FilterBuilder {
	b.Field = field
	return b
}

func (b *FilterBuilder) WithValue(value string) *FilterBuilder {
	b.Value = value
	return b
}

func (b *FilterBuilder) Build() Filter {
	return Filter{
		Field: b.Field,
		Value: b.Value,
	}
}

type SortCriteriaBuilder struct {
	Field SearchableField
	Order SortOrder
}

func NewSortCriteriaBuilder() *SortCriteriaBuilder {
	return &SortCriteriaBuilder{}
}

func (b *SortCriteriaBuilder) ForField(field SearchableField) *SortCriteriaBuilder {
	b.Field = field
	return b
}

func (b *SortCriteriaBuilder) Ascending() *SortCriteriaBuilder {
	b.Order = SortAsc
	return b
}

func (b *SortCriteriaBuilder) Descending() *SortCriteriaBuilder {
	b.Order = SortDesc
	return b
}

func (b *SortCriteriaBuilder) WithOrder(order SortOrder) *SortCriteriaBuilder {
	b.Order = order
	return b
}

func (b *SortCriteriaBuilder) Build() SortCriteria {
	return SortCriteria{
		Field: b.Field,
		Order: b.Order,
	}
}

func QuickSearchInput(
	searchTerm *string,
	pagination *cursor.CursorPaginationInput,
) (*ListSearchInput, error) {
	return NewListSearchInputBuilder().
		WithSearchTerm(searchTerm).
		WithSorting(FieldCreatedAt, SortDesc).
		WithPagination(pagination).
		Build()
}

func FilterOnlyInput(
	filters []Filter,
	pagination *cursor.CursorPaginationInput,
) (*ListSearchInput, error) {
	return NewListSearchInputBuilder().
		WithFilters(filters).
		WithSorting(FieldCreatedAt, SortDesc).
		WithPagination(pagination).
		Build()
}

func SearchWithSortInput(
	searchTerm *string,
	sorting []SortCriteria,
	pagination *cursor.CursorPaginationInput,
) (*ListSearchInput, error) {
	return NewListSearchInputBuilder().
		WithSearchTerm(searchTerm).
		WithSortingCriteria(sorting).
		WithPagination(pagination).
		Build()
}
