package cursor

import (
	"encoding/base64"
	"fmt"
)

func EncodeCursor(id string, sortValue interface{}, sortField string) string {
	var sortValueStr string
	switch v := sortValue.(type) {
	case string:
		sortValueStr = v
	case int, int32, int64:
		sortValueStr = fmt.Sprintf("%d", v)
	case float32, float64:
		sortValueStr = fmt.Sprintf("%f", v)
	default:
		sortValueStr = fmt.Sprintf("%v", v)
	}

	cursorContent := fmt.Sprintf("%s:%s:%s", sortField, sortValueStr, id)

	encoded := base64.StdEncoding.EncodeToString([]byte(cursorContent))
	return encoded
}

func EncodeCursorMultiSort(sortValues []interface{}) string {
	if len(sortValues) == 0 {
		return ""
	}

	parts := make([]string, len(sortValues))
	for i, val := range sortValues {
		switch v := val.(type) {
		case string:
			parts[i] = v
		case int, int32, int64:
			parts[i] = fmt.Sprintf("%d", v)
		case float32, float64:
			parts[i] = fmt.Sprintf("%f", v)
		default:
			parts[i] = fmt.Sprintf("%v", v)
		}
	}

	cursorContent := fmt.Sprintf("%v", parts)
	encoded := base64.StdEncoding.EncodeToString([]byte(cursorContent))
	return encoded
}
