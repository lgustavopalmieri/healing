package cursor

import (
	"encoding/base64"
	"fmt"
)

// EncodeCursor é uma função auxiliar para criar um cursor codificado
// a partir dos valores de um item.
//
// # Como usar:
//
// Esta função deve ser chamada pelo repository/service ao construir
// o CursorPaginationOutput. Ela pega os valores do último item da página
// e cria um cursor opaco que pode ser usado na próxima requisição.
//
// Exemplo de uso:
//
//	lastItem := items[len(items)-1]
//	nextCursor := cursor.EncodeCursor(
//	    lastItem.ID,
//	    lastItem.CreatedAt.Unix(), // valor usado para ordenação
//	    "created_at",               // nome do campo de ordenação
//	)
//
// # Formato do cursor:
//
// O cursor é codificado como: base64("sortField:sortValue:id")
// Exemplo: base64("created_at:1632489600:123") = "Y3JlYXRlZF9hdDoxNjMyNDg5NjAwOjEyMw=="
//
// Este formato permite:
//   - Ordenação consistente por qualquer campo
//   - Desempate por ID (garante unicidade)
//   - Fácil parsing no DecodeCursor
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
