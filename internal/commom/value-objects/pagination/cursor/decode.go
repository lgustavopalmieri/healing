package cursor

import (
	"encoding/base64"
	"strings"
)

type DecodedCursorValue struct {
	ID        string
	SortValue string
	SortField string
}

type DecodedMultiSortCursor struct {
	SortValues []interface{}
}

func (c *CursorPaginationInput) DecodeCursor() (*DecodedCursorValue, error) {
	if c.IsFirstPage() {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*c.EncodedCursor)
	if err != nil {
		return nil, ErrInvalidCursorFormat
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 3 {
		return nil, ErrInvalidCursorFormat
	}

	return &DecodedCursorValue{
		SortField: parts[0],
		SortValue: parts[1],
		ID:        parts[2],
	}, nil
}

func (c *CursorPaginationInput) DecodeMultiSortCursor() (*DecodedMultiSortCursor, error) {
	if c.IsFirstPage() {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*c.EncodedCursor)
	if err != nil {
		return nil, ErrInvalidCursorFormat
	}

	decodedStr := string(decoded)
	if !strings.HasPrefix(decodedStr, "[") || !strings.HasSuffix(decodedStr, "]") {
		return nil, ErrInvalidCursorFormat
	}

	content := strings.TrimPrefix(strings.TrimSuffix(decodedStr, "]"), "[")
	if content == "" {
		return &DecodedMultiSortCursor{SortValues: []interface{}{}}, nil
	}

	parts := strings.Split(content, " ")
	sortValues := make([]interface{}, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			sortValues = append(sortValues, part)
		}
	}

	return &DecodedMultiSortCursor{SortValues: sortValues}, nil
}
