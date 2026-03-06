package searchinput

import (
	"strings"
	"testing"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validPagination() *cursor.CursorPaginationInput {
	p, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
	return p
}

func strPtr(s string) *string {
	return &s
}

func TestNewListSearchInput(t *testing.T) {
	tests := []struct {
		name           string
		searchTerm     *string
		filters        []Filter
		sort           []Sort
		pagination     *cursor.CursorPaginationInput
		expectError    bool
		validateResult func(*testing.T, *ListSearchInput, error)
	}{
		{
			name:        "failure - returns ErrEmptySearchCriteria when no search term and no filters",
			searchTerm:  nil,
			filters:     nil,
			sort:        nil,
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				assert.Equal(t, search.ErrEmptySearchCriteria, err)
			},
		},
		{
			name:        "failure - returns ErrMissingPagination when pagination is nil",
			searchTerm:  strPtr("cardiologia"),
			filters:     nil,
			sort:        nil,
			pagination:  nil,
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				assert.Equal(t, search.ErrMissingPagination, err)
			},
		},
		{
			name:        "failure - returns ErrEmptySearchTerm when search term is empty string",
			searchTerm:  strPtr(""),
			filters:     []Filter{{Field: FieldSpecialty, Value: "Cardiologia"}},
			sort:        nil,
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				assert.Equal(t, search.ErrEmptySearchTerm, err)
			},
		},
		{
			name:        "failure - returns ErrEmptySearchTerm when search term is only whitespace",
			searchTerm:  strPtr("   "),
			filters:     []Filter{{Field: FieldSpecialty, Value: "Cardiologia"}},
			sort:        nil,
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				assert.Equal(t, search.ErrEmptySearchTerm, err)
			},
		},
		{
			name:        "failure - returns ErrSearchTermTooShort when search term has less than 2 chars",
			searchTerm:  strPtr("a"),
			filters:     nil,
			sort:        nil,
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				assert.Equal(t, search.ErrSearchTermTooShort, err)
			},
		},
		{
			name:        "failure - returns ErrSearchTermTooLong when search term exceeds 100 chars",
			searchTerm:  strPtr(strings.Repeat("a", 101)),
			filters:     nil,
			sort:        nil,
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				assert.Equal(t, search.ErrSearchTermTooLong, err)
			},
		},
		{
			name:        "success - creates input with valid search term only",
			searchTerm:  strPtr("cardiologia"),
			filters:     nil,
			sort:        nil,
			pagination:  validPagination(),
			expectError: false,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				require.NotNil(t, input)
				assert.Equal(t, "cardiologia", *input.SearchTerm)
				require.Len(t, input.Filters, 1)
				assert.Equal(t, FieldStatus, input.Filters[0].Field)
				assert.Equal(t, []string{"active", "authorized_license"}, input.Filters[0].Values)
				assert.NotEmpty(t, input.Sort)
			},
		},
		{
			name:       "success - creates input with filters only (no search term)",
			searchTerm: nil,
			filters: []Filter{
				{Field: FieldSpecialty, Value: "Cardiologia"},
			},
			sort:        nil,
			pagination:  validPagination(),
			expectError: false,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				require.NotNil(t, input)
				assert.Nil(t, input.SearchTerm)
				require.Len(t, input.Filters, 2)
				assert.Equal(t, FieldSpecialty, input.Filters[0].Field)
				assert.Equal(t, FieldStatus, input.Filters[1].Field)
				assert.Equal(t, []string{"active", "authorized_license"}, input.Filters[1].Values)
			},
		},
		{
			name:       "success - creates input with both search term and filters",
			searchTerm: strPtr("coração"),
			filters: []Filter{
				{Field: FieldSpecialty, Value: "Cardiologia"},
			},
			sort:        nil,
			pagination:  validPagination(),
			expectError: false,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				require.NotNil(t, input)
				assert.Equal(t, "coração", *input.SearchTerm)
				require.Len(t, input.Filters, 2)
				assert.Equal(t, FieldSpecialty, input.Filters[0].Field)
				assert.Equal(t, FieldStatus, input.Filters[1].Field)
				assert.Equal(t, []string{"active", "authorized_license"}, input.Filters[1].Values)
			},
		},
		{
			name:        "success - normalizes search term by trimming whitespace",
			searchTerm:  strPtr("  cardiologia  "),
			filters:     nil,
			sort:        nil,
			pagination:  validPagination(),
			expectError: false,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				require.NotNil(t, input)
				assert.Equal(t, "cardiologia", *input.SearchTerm)
			},
		},
		{
			name:       "success - normalizes filter values by trimming whitespace",
			searchTerm: strPtr("test"),
			filters: []Filter{
				{Field: FieldSpecialty, Value: "  Cardiologia  "},
			},
			sort:        nil,
			pagination:  validPagination(),
			expectError: false,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				require.NotNil(t, input)
				assert.Equal(t, "Cardiologia", input.Filters[0].Value)
			},
		},
		{
			name:       "failure - returns ErrInvalidSearchField when filter has invalid field",
			searchTerm: strPtr("test"),
			filters: []Filter{
				{Field: "invalid_field", Value: "value"},
			},
			sort:        nil,
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				var fieldErr *search.ErrInvalidSearchField
				assert.ErrorAs(t, err, &fieldErr)
				assert.Equal(t, "invalid_field", fieldErr.Field())
			},
		},
		{
			name:       "failure - returns ErrEmptyFilterValue when filter value is empty",
			searchTerm: strPtr("test"),
			filters: []Filter{
				{Field: FieldSpecialty, Value: "  "},
			},
			sort:        nil,
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				var filterErr *search.ErrEmptyFilterValue
				assert.ErrorAs(t, err, &filterErr)
				assert.Equal(t, "specialty", filterErr.Field())
			},
		},
		{
			name:       "failure - returns ErrDuplicateFilter when duplicate filter fields",
			searchTerm: strPtr("test"),
			filters: []Filter{
				{Field: FieldSpecialty, Value: "Cardiologia"},
				{Field: FieldSpecialty, Value: "Neurologia"},
			},
			sort:        nil,
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				var dupErr *search.ErrDuplicateFilter
				assert.ErrorAs(t, err, &dupErr)
				assert.Equal(t, "specialty", dupErr.Field())
			},
		},
		{
			name:       "failure - returns ErrInvalidSearchField when sort has invalid field",
			searchTerm: strPtr("test"),
			filters:    nil,
			sort: []Sort{
				{Field: "invalid_field", Order: SortAsc},
			},
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				var fieldErr *search.ErrInvalidSearchField
				assert.ErrorAs(t, err, &fieldErr)
			},
		},
		{
			name:       "failure - returns ErrFieldNotSortable when sort field is description",
			searchTerm: strPtr("test"),
			filters:    nil,
			sort: []Sort{
				{Field: FieldDescription, Order: SortAsc},
			},
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				var sortErr *search.ErrFieldNotSortable
				assert.ErrorAs(t, err, &sortErr)
				assert.Equal(t, "description", sortErr.Field())
			},
		},
		{
			name:       "failure - returns ErrFieldNotSortable when sort field is keywords",
			searchTerm: strPtr("test"),
			filters:    nil,
			sort: []Sort{
				{Field: FieldKeywords, Order: SortDesc},
			},
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				var sortErr *search.ErrFieldNotSortable
				assert.ErrorAs(t, err, &sortErr)
				assert.Equal(t, "keywords", sortErr.Field())
			},
		},
		{
			name:       "failure - returns ErrInvalidSortOrder when sort order is invalid",
			searchTerm: strPtr("test"),
			filters:    nil,
			sort: []Sort{
				{Field: FieldRating, Order: "invalid"},
			},
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				var orderErr *search.ErrInvalidSortOrder
				assert.ErrorAs(t, err, &orderErr)
				assert.Equal(t, "invalid", orderErr.Order())
			},
		},
		{
			name:       "failure - returns ErrDuplicateSortCriteria when duplicate sort fields",
			searchTerm: strPtr("test"),
			filters:    nil,
			sort: []Sort{
				{Field: FieldRating, Order: SortAsc},
				{Field: FieldRating, Order: SortDesc},
			},
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				var dupErr *search.ErrDuplicateSortCriteria
				assert.ErrorAs(t, err, &dupErr)
				assert.Equal(t, "rating", dupErr.Field())
			},
		},
		{
			name:       "failure - returns ErrFieldNotSupportsCursor when no sort field supports cursor",
			searchTerm: strPtr("test"),
			filters:    nil,
			sort: []Sort{
				{Field: FieldName, Order: SortAsc},
				{Field: FieldSpecialty, Order: SortDesc},
			},
			pagination:  validPagination(),
			expectError: true,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				assert.Nil(t, input)
				var cursorErr *search.ErrFieldNotSupportsCursor
				assert.ErrorAs(t, err, &cursorErr)
				assert.Equal(t, "name", cursorErr.Field())
			},
		},
		{
			name:       "success - allows sortable non-cursor field when cursor-compatible field also present",
			searchTerm: strPtr("test"),
			filters:    nil,
			sort: []Sort{
				{Field: FieldName, Order: SortAsc},
				{Field: FieldCreatedAt, Order: SortDesc},
			},
			pagination:  validPagination(),
			expectError: false,
			validateResult: func(t *testing.T, input *ListSearchInput, err error) {
				require.NotNil(t, input)
				hasName := false
				hasCreatedAt := false
				for _, s := range input.Sort {
					if s.Field == FieldName {
						hasName = true
					}
					if s.Field == FieldCreatedAt {
						hasCreatedAt = true
					}
				}
				assert.True(t, hasName)
				assert.True(t, hasCreatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := NewListSearchInput(tt.searchTerm, tt.filters, tt.sort, tt.pagination)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, input, err)
		})
	}
}

func TestEnsureDefaultSort(t *testing.T) {
	tests := []struct {
		name         string
		inputSort    []Sort
		validateSort func(*testing.T, []Sort)
	}{
		{
			name:      "adds rating and updated_at when no sort provided",
			inputSort: nil,
			validateSort: func(t *testing.T, sort []Sort) {
				require.Len(t, sort, 2)
				assert.Equal(t, FieldRating, sort[0].Field)
				assert.Equal(t, SortDesc, sort[0].Order)
				assert.Equal(t, FieldUpdatedAt, sort[1].Field)
				assert.Equal(t, SortDesc, sort[1].Order)
			},
		},
		{
			name: "does not duplicate rating when already present",
			inputSort: []Sort{
				{Field: FieldRating, Order: SortAsc},
			},
			validateSort: func(t *testing.T, sort []Sort) {
				require.Len(t, sort, 2)
				assert.Equal(t, FieldRating, sort[0].Field)
				assert.Equal(t, SortAsc, sort[0].Order)
				assert.Equal(t, FieldUpdatedAt, sort[1].Field)
				assert.Equal(t, SortDesc, sort[1].Order)
			},
		},
		{
			name: "does not duplicate updated_at when already present",
			inputSort: []Sort{
				{Field: FieldUpdatedAt, Order: SortAsc},
			},
			validateSort: func(t *testing.T, sort []Sort) {
				require.Len(t, sort, 2)
				assert.Equal(t, FieldUpdatedAt, sort[0].Field)
				assert.Equal(t, SortAsc, sort[0].Order)
				assert.Equal(t, FieldRating, sort[1].Field)
				assert.Equal(t, SortDesc, sort[1].Order)
			},
		},
		{
			name: "preserves custom sort and adds defaults",
			inputSort: []Sort{
				{Field: FieldRating, Order: SortDesc},
				{Field: FieldUpdatedAt, Order: SortDesc},
			},
			validateSort: func(t *testing.T, sort []Sort) {
				require.Len(t, sort, 2)
				assert.Equal(t, FieldRating, sort[0].Field)
				assert.Equal(t, FieldUpdatedAt, sort[1].Field)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchTerm := "test"
			pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)

			input, err := NewListSearchInput(&searchTerm, nil, tt.inputSort, pagination)

			require.NoError(t, err)
			tt.validateSort(t, input.Sort)
		})
	}
}
