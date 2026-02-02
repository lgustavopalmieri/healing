package cursor

// CursorPaginationInput é um Value Object que representa os parâmetros de entrada
// para paginação baseada em cursor.
//
// # O que é Paginação por Cursor?
//
// Paginação por cursor é uma técnica onde, ao invés de usar números de página (1, 2, 3...),
// usamos um "ponteiro" (cursor) que aponta para um registro específico no conjunto de dados.
// O cursor geralmente contém informações sobre o último item visualizado, permitindo
// buscar os próximos itens a partir daquele ponto.
//
// # Vantagens sobre paginação offset/limit:
//   - Consistência: Mesmo se novos itens forem adicionados, você não verá duplicatas
//   - Performance: Não precisa contar ou pular registros (OFFSET é lento em grandes datasets)
//   - Adequado para feeds infinitos e dados em tempo real
//
// # Como usar este Value Object:
//
//	// Primeira requisição (sem cursor)
//	input, err := cursor.NewCursorPaginationInput(nil, 20, DirectionNext)
//	if err != nil {
//	    // tratar erro de validação
//	}
//
//	// Requisições subsequentes (com cursor da resposta anterior)
//	nextCursor := "eyJpZCI6MTIzLCJ0aW1lc3RhbXAiOjE2MzI0..." // vindo do output anterior
//	input, err := cursor.NewCursorPaginationInput(&nextCursor, 20, DirectionNext)
//
// # Estrutura do Cursor:
//
// O cursor é uma string opaca (geralmente base64) que contém informações sobre
// a posição atual na lista. O cliente não deve interpretar ou modificar o cursor,
// apenas armazená-lo e enviá-lo de volta nas próximas requisições.
type CursorPaginationInput struct {
	// EncodedCursor é o cursor codificado que aponta para uma posição específica
	// no conjunto de dados. Nil ou string vazia indica o início do conjunto.
	//
	// Exemplo de cursor: "eyJpZCI6MTIzLCJ0aW1lc3RhbXAiOjE2MzI0ODk2MDB9"
	// (base64 de: {"id":123,"timestamp":1632489600})
	EncodedCursor *string

	// PageSize define quantos itens devem ser retornados por página.
	// Deve ser sempre maior que zero.
	PageSize int

	// Direction indica se estamos navegando para frente (próximos itens)
	// ou para trás (itens anteriores).
	Direction PaginationDirection
}

// PaginationDirection representa a direção da navegação na paginação.
//
// Em paginação por cursor, você pode navegar em duas direções:
//   - DirectionNext: busca os próximos N itens após o cursor atual
//   - DirectionPrevious: busca os N itens anteriores ao cursor atual
//
// Isso permite implementar navegação bidirecional (avançar e voltar).
type PaginationDirection string

const (
	// DirectionNext indica que queremos os próximos itens após o cursor.
	// Exemplo: Se o cursor aponta para o item #10, buscar os próximos 5 itens (#11-#15)
	DirectionNext PaginationDirection = "next"

	// DirectionPrevious indica que queremos os itens anteriores ao cursor.
	// Exemplo: Se o cursor aponta para o item #10, buscar os 5 itens anteriores (#5-#9)
	DirectionPrevious PaginationDirection = "previous"
)

// NewCursorPaginationInput cria e valida um novo input de paginação por cursor.
//
// Parâmetros:
//   - encodedCursor: ponteiro para string com o cursor (nil para primeira página)
//   - pageSize: quantidade de itens por página (deve ser > 0)
//   - direction: direção da navegação (next ou previous)
//
// Retorna erro se alguma validação de domínio falhar.
//
// Exemplo de uso:
//
//	// Primeira página
//	input, err := cursor.NewCursorPaginationInput(nil, 20, cursor.DirectionNext)
//
//	// Próxima página
//	nextCursor := "abc123..."
//	input, err := cursor.NewCursorPaginationInput(&nextCursor, 20, cursor.DirectionNext)
//
//	// Página anterior
//	prevCursor := "xyz789..."
//	input, err := cursor.NewCursorPaginationInput(&prevCursor, 20, cursor.DirectionPrevious)
func NewCursorPaginationInput(
	encodedCursor *string,
	pageSize int,
	direction PaginationDirection,
) (*CursorPaginationInput, error) {
	input := &CursorPaginationInput{
		EncodedCursor: encodedCursor,
		PageSize:      pageSize,
		Direction:     direction,
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	return input, nil
}

// IsFirstPage indica se esta é a primeira página (sem cursor).
func (c *CursorPaginationInput) IsFirstPage() bool {
	return c.EncodedCursor == nil || *c.EncodedCursor == ""
}

// IsNavigatingForward indica se estamos navegando para frente.
func (c *CursorPaginationInput) IsNavigatingForward() bool {
	return c.Direction == DirectionNext
}

// IsNavigatingBackward indica se estamos navegando para trás.
func (c *CursorPaginationInput) IsNavigatingBackward() bool {
	return c.Direction == DirectionPrevious
}
