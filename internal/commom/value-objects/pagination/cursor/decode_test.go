package cursor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeMultiSortCursor(t *testing.T) {
	tests := []struct {
		name        string
		setupCursor func() *string
		expectError bool
		validate    func(*testing.T, *DecodedMultiSortCursor)
	}{
		{
			name: "decodes multi-sort cursor correctly",
			setupCursor: func() *string {
				encoded := EncodeCursorMultiSort([]interface{}{4.5, "2024-01-15T10:30:00Z", "uuid-123"})
				return &encoded
			},
			expectError: false,
			validate: func(t *testing.T, decoded *DecodedMultiSortCursor) {
				require.NotNil(t, decoded)
				assert.Len(t, decoded.SortValues, 3)
			},
		},
		{
			name: "returns nil for first page",
			setupCursor: func() *string {
				return nil
			},
			expectError: false,
			validate: func(t *testing.T, decoded *DecodedMultiSortCursor) {
				assert.Nil(t, decoded)
			},
		},
		{
			name: "returns error for invalid base64",
			setupCursor: func() *string {
				invalid := "not-valid-base64!!!"
				return &invalid
			},
			expectError: true,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := tt.setupCursor()
			input := &CursorPaginationInput{
				EncodedCursor: cursor,
				PageSize:      10,
				Direction:     DirectionNext,
			}

			decoded, err := input.DecodeMultiSortCursor()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, decoded)
				}
			}
		})
	}
}
