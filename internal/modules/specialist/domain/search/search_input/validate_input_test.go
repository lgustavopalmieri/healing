package searchinput

import (
	"testing"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
