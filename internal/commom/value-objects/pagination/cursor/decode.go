package cursor

import (
	"encoding/base64"
	"strings"
)

// DecodedCursorValue representa o conteúdo decodificado de um cursor.
//
// # Estrutura do Cursor
//
// Um cursor contém informações que permitem localizar exatamente onde
// parar a busca anterior e onde começar a próxima. Tipicamente contém:
//   - ID do último item visualizado
//   - Timestamp ou campo de ordenação
//   - Qualquer outro campo necessário para ordenação única
//
// Exemplo: Se você ordena por "created_at DESC, id DESC", o cursor deve
// conter ambos os valores para garantir ordenação consistente.
type DecodedCursorValue struct {
	// ID é o identificador único do último item visualizado.
	// Usado como fallback para garantir ordenação única.
	ID string

	// SortValue é o valor do campo usado para ordenação.
	// Exemplo: timestamp, score, nome, etc.
	SortValue string

	// SortField indica qual campo está sendo usado para ordenação.
	// Exemplo: "created_at", "updated_at", "score", etc.
	SortField string
}

// DecodeCursor decodifica o cursor de base64 para seus valores originais.
//
// O formato esperado é: "sortField:sortValue:id"
// Exemplo: "created_at:1632489600:123"
//
// Retorna erro se o cursor estiver em formato inválido.
func (c *CursorPaginationInput) DecodeCursor() (*DecodedCursorValue, error) {
	if c.IsFirstPage() {
		return nil, nil
	}

	// Decodifica de base64
	decoded, err := base64.StdEncoding.DecodeString(*c.EncodedCursor)
	if err != nil {
		return nil, ErrInvalidCursorFormat
	}

	// Parse do formato: "sortField:sortValue:id"
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
