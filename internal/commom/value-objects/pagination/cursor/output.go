package cursor

// CursorPaginationOutput é um Value Object que representa o resultado
// de uma operação de paginação por cursor.
//
// # Como interpretar este output:
//
// Este objeto contém todas as informações necessárias para o cliente
// implementar navegação bidirecional (avançar e voltar):
//
//   - NextCursor: use este cursor para buscar a próxima página
//   - PreviousCursor: use este cursor para voltar à página anterior
//   - HasNextPage: indica se existem mais itens após a página atual
//   - HasPreviousPage: indica se existem itens antes da página atual
//
// # Exemplo de uso no cliente:
//
//	// Recebeu o output da primeira requisição
//	output := service.ListItems(input)
//
//	// Verificar se há próxima página
//	if output.HasNextPage() {
//	    nextCursor := output.NextCursor()
//	    // Fazer nova requisição com nextCursor
//	    nextInput, _ := cursor.NewCursorPaginationInput(nextCursor, 20, cursor.DirectionNext)
//	    nextOutput := service.ListItems(nextInput)
//	}
//
//	// Verificar se pode voltar
//	if output.HasPreviousPage() {
//	    prevCursor := output.PreviousCursor()
//	    // Fazer nova requisição com prevCursor
//	    prevInput, _ := cursor.NewCursorPaginationInput(prevCursor, 20, cursor.DirectionPrevious)
//	    prevOutput := service.ListItems(prevInput)
//	}
//
// # Importante:
//
// Os cursores são strings opacas. O cliente NUNCA deve:
//   - Tentar decodificar ou interpretar o cursor
//   - Modificar o cursor
//   - Assumir qualquer estrutura interna
//
// O cliente deve apenas:
//   - Armazenar o cursor recebido
//   - Enviar o cursor de volta nas próximas requisições
//   - Usar os flags HasNextPage/HasPreviousPage para controlar navegação
type CursorPaginationOutput struct {
	// NextCursor é o cursor para buscar a próxima página.
	// Nil se não houver próxima página.
	NextCursor *string

	// PreviousCursor é o cursor para buscar a página anterior.
	// Nil se não houver página anterior.
	PreviousCursor *string

	// HasNextPage indica se existem mais itens após esta página.
	HasNextPage bool

	// HasPreviousPage indica se existem itens antes desta página.
	HasPreviousPage bool

	// TotalItemsInPage indica quantos itens foram retornados nesta página.
	// Útil para o cliente saber se a página está parcialmente preenchida.
	TotalItemsInPage int
}

// NewCursorPaginationOutput cria um novo output de paginação por cursor.
//
// Parâmetros:
//   - nextCursor: cursor para próxima página (nil se não houver)
//   - previousCursor: cursor para página anterior (nil se não houver)
//   - hasNextPage: indica se há mais itens após esta página
//   - hasPreviousPage: indica se há itens antes desta página
//   - totalItemsInPage: quantidade de itens retornados nesta página
//
// Exemplo de uso no repository/service:
//
//	// Buscar N+1 itens para saber se há próxima página
//	items := repository.FindItems(cursor, pageSize+1)
//
//	hasNext := len(items) > pageSize
//	if hasNext {
//	    items = items[:pageSize] // remover o item extra
//	}
//
//	// Gerar cursores
//	var nextCursor *string
//	if hasNext {
//	    lastItem := items[len(items)-1]
//	    encoded := encodeCursor(lastItem)
//	    nextCursor = &encoded
//	}
//
//	var prevCursor *string
//	if !isFirstPage {
//	    firstItem := items[0]
//	    encoded := encodeCursor(firstItem)
//	    prevCursor = &encoded
//	}
//
//	output := cursor.NewCursorPaginationOutput(
//	    nextCursor,
//	    prevCursor,
//	    hasNext,
//	    !isFirstPage,
//	    len(items),
//	)
func NewCursorPaginationOutput(
	nextCursor *string,
	previousCursor *string,
	hasNextPage bool,
	hasPreviousPage bool,
	totalItemsInPage int,
) *CursorPaginationOutput {
	return &CursorPaginationOutput{
		NextCursor:       nextCursor,
		PreviousCursor:   previousCursor,
		HasNextPage:      hasNextPage,
		HasPreviousPage:  hasPreviousPage,
		TotalItemsInPage: totalItemsInPage,
	}
}

// IsEmpty indica se esta página não contém nenhum item.
func (c *CursorPaginationOutput) IsEmpty() bool {
	return c.TotalItemsInPage == 0
}

// IsPartialPage indica se esta página contém menos itens que o solicitado.
// Geralmente indica que é a última página.
func (c *CursorPaginationOutput) IsPartialPage(requestedPageSize int) bool {
	return c.TotalItemsInPage < requestedPageSize && c.TotalItemsInPage > 0
}
