package cursor

import "strconv"

// ParseIntFromCursor é uma função auxiliar para extrair um valor inteiro
// do cursor decodificado.
//
// Útil quando o sortValue é um timestamp ou ID numérico.
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
