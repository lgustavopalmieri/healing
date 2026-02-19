package cursor

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeCursorMultiSort(t *testing.T) {
	tests := []struct {
		name       string
		sortValues []interface{}
		validate   func(*testing.T, string)
	}{
		{
			name:       "encodes multiple sort values correctly",
			sortValues: []interface{}{4.5, "2024-01-15T10:30:00Z", "uuid-123"},
			validate: func(t *testing.T, encoded string) {
				assert.NotEmpty(t, encoded)
				decoded, err := base64.StdEncoding.DecodeString(encoded)
				assert.NoError(t, err)
				assert.Contains(t, string(decoded), "4.5")
				assert.Contains(t, string(decoded), "2024-01-15T10:30:00Z")
				assert.Contains(t, string(decoded), "uuid-123")
			},
		},
		{
			name:       "handles empty slice",
			sortValues: []interface{}{},
			validate: func(t *testing.T, encoded string) {
				assert.Empty(t, encoded)
			},
		},
		{
			name:       "handles different types",
			sortValues: []interface{}{5.0, int64(1234567890), "id-456"},
			validate: func(t *testing.T, encoded string) {
				assert.NotEmpty(t, encoded)
				decoded, err := base64.StdEncoding.DecodeString(encoded)
				assert.NoError(t, err)
				assert.Contains(t, string(decoded), "5.0")
				assert.Contains(t, string(decoded), "1234567890")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeCursorMultiSort(tt.sortValues)
			tt.validate(t, encoded)
		})
	}
}
