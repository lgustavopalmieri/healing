package cursor

import "strconv"

func ParseIntFromCursor(decoded *DecodedCursorValue) (int64, error) {
	if decoded == nil {
		return 0, ErrInvalidCursorFormat
	}

	value, err := strconv.ParseInt(decoded.SortValue, 10, 64)
	if err != nil {
		return 0, ErrInvalidCursorFormat
	}

	return value, nil
}
