package cursor

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

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
	// encodedCursor é o cursor codificado que aponta para uma posição específica
	// no conjunto de dados. Nil ou string vazia indica o início do conjunto.
	//
	// Exemplo de cursor: "eyJpZCI6MTIzLCJ0aW1lc3RhbXAiOjE2MzI0ODk2MDB9"
	// (base64 de: {"id":123,"timestamp":1632489600})
	encodedCursor *string

	// pageSize define quantos itens devem ser retornados por página.
	// Deve ser sempre maior que zero.
	pageSize int

	// direction indica se estamos navegando para frente (próximos itens)
	// ou para trás (itens anteriores).
	direction PaginationDirection
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
		encodedCursor: encodedCursor,
		pageSize:      pageSize,
		direction:     direction,
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	return input, nil
}

// validate executa todas as validações de domínio para o input de paginação.
func (c *CursorPaginationInput) validate() error {
	if err := c.validatePageSize(); err != nil {
		return err
	}

	if err := c.validateDirection(); err != nil {
		return err
	}

	if err := c.validateCursor(); err != nil {
		return err
	}

	return nil
}

// validatePageSize garante que o tamanho da página é válido.
func (c *CursorPaginationInput) validatePageSize() error {
	if c.pageSize <= 0 {
		return ErrInvalidPageSize
	}

	// Opcional: você pode adicionar um limite máximo para evitar sobrecarga
	const maxPageSize = 100
	if c.pageSize > maxPageSize {
		return ErrPageSizeTooLarge
	}

	return nil
}

// validateDirection garante que a direção é válida.
func (c *CursorPaginationInput) validateDirection() error {
	if c.direction != DirectionNext && c.direction != DirectionPrevious {
		return ErrInvalidDirection
	}
	return nil
}

// validateCursor valida o formato do cursor, se fornecido.
func (c *CursorPaginationInput) validateCursor() error {
	// Cursor nil ou vazio é válido (indica primeira página)
	if c.encodedCursor == nil || *c.encodedCursor == "" {
		return nil
	}

	// Valida se o cursor está em formato base64 válido
	// (isso é uma validação básica - a validação completa acontece ao decodificar)
	_, err := base64.StdEncoding.DecodeString(*c.encodedCursor)
	if err != nil {
		return ErrInvalidCursorFormat
	}

	return nil
}

// EncodedCursor retorna o cursor codificado.
// Retorna nil se não há cursor (primeira página).
func (c *CursorPaginationInput) EncodedCursor() *string {
	return c.encodedCursor
}

// PageSize retorna o tamanho da página solicitado.
func (c *CursorPaginationInput) PageSize() int {
	return c.pageSize
}

// Direction retorna a direção da navegação.
func (c *CursorPaginationInput) Direction() PaginationDirection {
	return c.direction
}

// IsFirstPage indica se esta é a primeira página (sem cursor).
func (c *CursorPaginationInput) IsFirstPage() bool {
	return c.encodedCursor == nil || *c.encodedCursor == ""
}

// IsNavigatingForward indica se estamos navegando para frente.
func (c *CursorPaginationInput) IsNavigatingForward() bool {
	return c.direction == DirectionNext
}

// IsNavigatingBackward indica se estamos navegando para trás.
func (c *CursorPaginationInput) IsNavigatingBackward() bool {
	return c.direction == DirectionPrevious
}

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
	decoded, err := base64.StdEncoding.DecodeString(*c.encodedCursor)
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
	// nextCursor é o cursor para buscar a próxima página.
	// Nil se não houver próxima página.
	nextCursor *string

	// previousCursor é o cursor para buscar a página anterior.
	// Nil se não houver página anterior.
	previousCursor *string

	// hasNextPage indica se existem mais itens após esta página.
	hasNextPage bool

	// hasPreviousPage indica se existem itens antes desta página.
	hasPreviousPage bool

	// totalItemsInPage indica quantos itens foram retornados nesta página.
	// Útil para o cliente saber se a página está parcialmente preenchida.
	totalItemsInPage int
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
		nextCursor:       nextCursor,
		previousCursor:   previousCursor,
		hasNextPage:      hasNextPage,
		hasPreviousPage:  hasPreviousPage,
		totalItemsInPage: totalItemsInPage,
	}
}

// NextCursor retorna o cursor para buscar a próxima página.
// Retorna nil se não houver próxima página.
//
// Use este cursor em uma nova requisição com Direction = DirectionNext.
func (c *CursorPaginationOutput) NextCursor() *string {
	return c.nextCursor
}

// PreviousCursor retorna o cursor para buscar a página anterior.
// Retorna nil se não houver página anterior.
//
// Use este cursor em uma nova requisição com Direction = DirectionPrevious.
func (c *CursorPaginationOutput) PreviousCursor() *string {
	return c.previousCursor
}

// HasNextPage indica se existem mais itens após esta página.
//
// Use este método para:
//   - Habilitar/desabilitar botão "Próxima" na UI
//   - Implementar scroll infinito
//   - Decidir se deve pré-carregar próxima página
func (c *CursorPaginationOutput) HasNextPage() bool {
	return c.hasNextPage
}

// HasPreviousPage indica se existem itens antes desta página.
//
// Use este método para:
//   - Habilitar/desabilitar botão "Anterior" na UI
//   - Implementar navegação bidirecional
func (c *CursorPaginationOutput) HasPreviousPage() bool {
	return c.hasPreviousPage
}

// TotalItemsInPage retorna quantos itens foram retornados nesta página.
//
// Útil para:
//   - Mostrar "Exibindo X itens" na UI
//   - Detectar última página (se < pageSize solicitado)
//   - Métricas e analytics
func (c *CursorPaginationOutput) TotalItemsInPage() int {
	return c.totalItemsInPage
}

// IsEmpty indica se esta página não contém nenhum item.
func (c *CursorPaginationOutput) IsEmpty() bool {
	return c.totalItemsInPage == 0
}

// IsPartialPage indica se esta página contém menos itens que o solicitado.
// Geralmente indica que é a última página.
func (c *CursorPaginationOutput) IsPartialPage(requestedPageSize int) bool {
	return c.totalItemsInPage < requestedPageSize && c.totalItemsInPage > 0
}

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
	// Converte sortValue para string
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

	// Formato: "sortField:sortValue:id"
	cursorContent := fmt.Sprintf("%s:%s:%s", sortField, sortValueStr, id)

	// Codifica em base64
	encoded := base64.StdEncoding.EncodeToString([]byte(cursorContent))
	return encoded
}

// CursorBuilder é um helper para construir CursorPaginationOutput
// de forma mais fluente e legível.
//
// # Exemplo de uso:
//
//	output := cursor.NewCursorBuilder().
//	    WithNextCursor(nextCursor).
//	    WithPreviousCursor(prevCursor).
//	    WithHasNextPage(true).
//	    WithHasPreviousPage(false).
//	    WithTotalItems(20).
//	    Build()
type CursorBuilder struct {
	nextCursor       *string
	previousCursor   *string
	hasNextPage      bool
	hasPreviousPage  bool
	totalItemsInPage int
}

// NewCursorBuilder cria um novo builder para CursorPaginationOutput.
func NewCursorBuilder() *CursorBuilder {
	return &CursorBuilder{}
}

// WithNextCursor define o cursor para próxima página.
func (b *CursorBuilder) WithNextCursor(cursor *string) *CursorBuilder {
	b.nextCursor = cursor
	return b
}

// WithPreviousCursor define o cursor para página anterior.
func (b *CursorBuilder) WithPreviousCursor(cursor *string) *CursorBuilder {
	b.previousCursor = cursor
	return b
}

// WithHasNextPage define se há próxima página.
func (b *CursorBuilder) WithHasNextPage(hasNext bool) *CursorBuilder {
	b.hasNextPage = hasNext
	return b
}

// WithHasPreviousPage define se há página anterior.
func (b *CursorBuilder) WithHasPreviousPage(hasPrev bool) *CursorBuilder {
	b.hasPreviousPage = hasPrev
	return b
}

// WithTotalItems define o total de itens na página.
func (b *CursorBuilder) WithTotalItems(total int) *CursorBuilder {
	b.totalItemsInPage = total
	return b
}

// Build constrói o CursorPaginationOutput final.
func (b *CursorBuilder) Build() *CursorPaginationOutput {
	return NewCursorPaginationOutput(
		b.nextCursor,
		b.previousCursor,
		b.hasNextPage,
		b.hasPreviousPage,
		b.totalItemsInPage,
	)
}

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
