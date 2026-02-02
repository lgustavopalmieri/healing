package search

import "fmt"

// Domain errors para busca avançada de especialistas.
//
// Estes erros representam violações de regras de negócio relacionadas
// à busca avançada. São erros de domínio, não erros técnicos.

var (
	// ErrEmptySearchCriteria indica que nenhum critério de busca foi fornecido.
	//
	// Regra de negócio violada: Uma busca deve ter pelo menos um critério
	// (searchTerm ou filters). Buscas sem critérios retornariam todos os
	// registros, o que não é semânticamente uma "busca".
	//
	// Como tratar: Solicitar ao cliente que forneça pelo menos um searchTerm
	// ou um filtro.
	ErrEmptySearchCriteria = NewDomainError(
		"EMPTY_SEARCH_CRITERIA",
		"a busca deve conter pelo menos um critério (termo de busca ou filtros)",
	)

	// ErrMissingPagination indica que os parâmetros de paginação não foram fornecidos.
	//
	// Regra de negócio violada: Toda busca deve ter paginação para evitar
	// retornar conjuntos de dados muito grandes.
	//
	// Como tratar: Sempre fornecer parâmetros de paginação válidos.
	ErrMissingPagination = NewDomainError(
		"MISSING_PAGINATION",
		"os parâmetros de paginação são obrigatórios",
	)

	// ErrEmptySearchTerm indica que o termo de busca está vazio.
	//
	// Regra de negócio violada: Se um searchTerm for fornecido, ele não pode
	// ser uma string vazia ou apenas espaços em branco.
	//
	// Como tratar: Remover o searchTerm ou fornecer um valor válido.
	ErrEmptySearchTerm = NewDomainError(
		"EMPTY_SEARCH_TERM",
		"o termo de busca não pode ser vazio",
	)

	// ErrSearchTermTooShort indica que o termo de busca é muito curto.
	//
	// Regra de negócio violada: Termos de busca muito curtos (< 2 caracteres)
	// resultam em buscas muito genéricas e podem causar problemas de performance.
	//
	// Como tratar: Fornecer um termo de busca com pelo menos 2 caracteres.
	ErrSearchTermTooShort = NewDomainError(
		"SEARCH_TERM_TOO_SHORT",
		"o termo de busca deve ter pelo menos 2 caracteres",
	)

	// ErrSearchTermTooLong indica que o termo de busca é muito longo.
	//
	// Regra de negócio violada: Termos de busca muito longos (> 100 caracteres)
	// podem indicar tentativa de ataque ou uso incorreto da API.
	//
	// Como tratar: Fornecer um termo de busca com no máximo 100 caracteres.
	ErrSearchTermTooLong = NewDomainError(
		"SEARCH_TERM_TOO_LONG",
		"o termo de busca não pode exceder 100 caracteres",
	)
)

// ErrInvalidSearchField indica que um campo de busca inválido foi usado.
type ErrInvalidSearchField struct {
	*DomainError
	field string
}

// NewErrInvalidSearchField cria um novo erro de campo de busca inválido.
func NewErrInvalidSearchField(field string) *ErrInvalidSearchField {
	return &ErrInvalidSearchField{
		DomainError: NewDomainError(
			"INVALID_SEARCH_FIELD",
			fmt.Sprintf("o campo '%s' não é um campo de busca válido", field),
		),
		field: field,
	}
}

// Field retorna o campo inválido.
func (e *ErrInvalidSearchField) Field() string {
	return e.field
}

// ErrFieldNotFilterable indica que um campo não pode ser usado em filtros.
type ErrFieldNotFilterable struct {
	*DomainError
	field string
}

// NewErrFieldNotFilterable cria um novo erro de campo não filtrável.
func NewErrFieldNotFilterable(field string) *ErrFieldNotFilterable {
	return &ErrFieldNotFilterable{
		DomainError: NewDomainError(
			"FIELD_NOT_FILTERABLE",
			fmt.Sprintf("o campo '%s' não pode ser usado em filtros exatos (use busca textual)", field),
		),
		field: field,
	}
}

// Field retorna o campo não filtrável.
func (e *ErrFieldNotFilterable) Field() string {
	return e.field
}

// ErrFieldNotSortable indica que um campo não pode ser usado para ordenação.
type ErrFieldNotSortable struct {
	*DomainError
	field string
}

// NewErrFieldNotSortable cria um novo erro de campo não ordenável.
func NewErrFieldNotSortable(field string) *ErrFieldNotSortable {
	return &ErrFieldNotSortable{
		DomainError: NewDomainError(
			"FIELD_NOT_SORTABLE",
			fmt.Sprintf("o campo '%s' não pode ser usado para ordenação", field),
		),
		field: field,
	}
}

// Field retorna o campo não ordenável.
func (e *ErrFieldNotSortable) Field() string {
	return e.field
}

// ErrFieldNotSupportsCursor indica que um campo não suporta paginação por cursor.
type ErrFieldNotSupportsCursor struct {
	*DomainError
	field string
}

// NewErrFieldNotSupportsCursor cria um novo erro de campo que não suporta cursor.
func NewErrFieldNotSupportsCursor(field string) *ErrFieldNotSupportsCursor {
	return &ErrFieldNotSupportsCursor{
		DomainError: NewDomainError(
			"FIELD_NOT_SUPPORTS_CURSOR",
			fmt.Sprintf("o campo '%s' não pode ser usado como critério primário de ordenação em paginação por cursor (use created_at ou updated_at)", field),
		),
		field: field,
	}
}

// Field retorna o campo que não suporta cursor.
func (e *ErrFieldNotSupportsCursor) Field() string {
	return e.field
}

// ErrEmptyFilterValue indica que o valor de um filtro está vazio.
type ErrEmptyFilterValue struct {
	*DomainError
	field string
}

// NewErrEmptyFilterValue cria um novo erro de valor de filtro vazio.
func NewErrEmptyFilterValue(field string) *ErrEmptyFilterValue {
	return &ErrEmptyFilterValue{
		DomainError: NewDomainError(
			"EMPTY_FILTER_VALUE",
			fmt.Sprintf("o valor do filtro para o campo '%s' não pode ser vazio", field),
		),
		field: field,
	}
}

// Field retorna o campo com valor vazio.
func (e *ErrEmptyFilterValue) Field() string {
	return e.field
}

// ErrDuplicateFilter indica que há filtros duplicados para o mesmo campo.
type ErrDuplicateFilter struct {
	*DomainError
	field string
}

// NewErrDuplicateFilter cria um novo erro de filtro duplicado.
func NewErrDuplicateFilter(field string) *ErrDuplicateFilter {
	return &ErrDuplicateFilter{
		DomainError: NewDomainError(
			"DUPLICATE_FILTER",
			fmt.Sprintf("não é permitido ter múltiplos filtros para o campo '%s'", field),
		),
		field: field,
	}
}

// Field retorna o campo duplicado.
func (e *ErrDuplicateFilter) Field() string {
	return e.field
}

// ErrDuplicateSortCriteria indica que há critérios de ordenação duplicados.
type ErrDuplicateSortCriteria struct {
	*DomainError
	field string
}

// NewErrDuplicateSortCriteria cria um novo erro de critério de ordenação duplicado.
func NewErrDuplicateSortCriteria(field string) *ErrDuplicateSortCriteria {
	return &ErrDuplicateSortCriteria{
		DomainError: NewDomainError(
			"DUPLICATE_SORT_CRITERIA",
			fmt.Sprintf("não é permitido ter múltiplos critérios de ordenação para o campo '%s'", field),
		),
		field: field,
	}
}

// Field retorna o campo duplicado.
func (e *ErrDuplicateSortCriteria) Field() string {
	return e.field
}

// ErrInvalidSortOrder indica que a ordem de ordenação é inválida.
type ErrInvalidSortOrder struct {
	*DomainError
	order string
}

// NewErrInvalidSortOrder cria um novo erro de ordem de ordenação inválida.
func NewErrInvalidSortOrder(order string) *ErrInvalidSortOrder {
	return &ErrInvalidSortOrder{
		DomainError: NewDomainError(
			"INVALID_SORT_ORDER",
			fmt.Sprintf("a ordem de ordenação '%s' é inválida (use 'asc' ou 'desc')", order),
		),
		order: order,
	}
}

// Order retorna a ordem inválida.
func (e *ErrInvalidSortOrder) Order() string {
	return e.order
}

// DomainError é a estrutura base para erros de domínio.
type DomainError struct {
	code    string
	message string
}

// NewDomainError cria um novo erro de domínio.
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		code:    code,
		message: message,
	}
}

// Error implementa a interface error.
func (e *DomainError) Error() string {
	return e.message
}

// Code retorna o código do erro.
func (e *DomainError) Code() string {
	return e.code
}

// Message retorna a mensagem do erro.
func (e *DomainError) Message() string {
	return e.message
}

// IsListSearchDomainError verifica se um erro é um erro de domínio de busca.
//
// Útil para distinguir erros de domínio de erros técnicos (banco de dados,
// rede, etc.) e tratá-los de forma diferente na camada de aplicação.
//
// Exemplo de uso:
//
//	if list.IsListSearchDomainError(err) {
//	    // Retornar 400 Bad Request
//	    return BadRequestResponse(err)
//	}
//	// Retornar 500 Internal Server Error
//	return InternalErrorResponse()
func IsListSearchDomainError(err error) bool {
	switch err.(type) {
	case *DomainError,
		*ErrInvalidSearchField,
		*ErrFieldNotFilterable,
		*ErrFieldNotSortable,
		*ErrFieldNotSupportsCursor,
		*ErrEmptyFilterValue,
		*ErrDuplicateFilter,
		*ErrDuplicateSortCriteria,
		*ErrInvalidSortOrder:
		return true
	default:
		return false
	}
}
