package cursor

import "errors"

// Domain errors para paginação por cursor.
//
// Estes erros representam violações de regras de negócio relacionadas
// à paginação. São erros de domínio, não erros técnicos.

var (
	// ErrInvalidPageSize indica que o tamanho da página solicitado é inválido.
	//
	// Regra de negócio violada: O tamanho da página deve ser maior que zero.
	//
	// Exemplo de violação:
	//   - pageSize = 0
	//   - pageSize = -10
	//
	// Como tratar: Solicitar ao cliente que envie um pageSize válido (> 0).
	ErrInvalidPageSize = errors.New("o tamanho da página deve ser maior que zero")

	// ErrPageSizeTooLarge indica que o tamanho da página excede o limite permitido.
	//
	// Regra de negócio violada: O tamanho da página não pode exceder o máximo
	// configurado (geralmente 100 itens) para evitar sobrecarga no sistema.
	//
	// Exemplo de violação:
	//   - pageSize = 1000 (quando o máximo é 100)
	//
	// Como tratar: Informar ao cliente o limite máximo permitido e solicitar
	// um valor dentro do range aceitável.
	ErrPageSizeTooLarge = errors.New("o tamanho da página excede o limite máximo permitido (100 itens)")

	// ErrInvalidDirection indica que a direção de navegação é inválida.
	//
	// Regra de negócio violada: A direção deve ser "next" ou "previous".
	//
	// Exemplo de violação:
	//   - direction = "forward"
	//   - direction = "back"
	//   - direction = ""
	//
	// Como tratar: Solicitar ao cliente que use apenas "next" ou "previous".
	ErrInvalidDirection = errors.New("direção de paginação inválida: use 'next' ou 'previous'")

	// ErrInvalidCursorFormat indica que o cursor fornecido está em formato inválido.
	//
	// Regra de negócio violada: O cursor deve ser uma string base64 válida
	// contendo dados estruturados no formato esperado.
	//
	// Possíveis causas:
	//   - Cursor não está em base64
	//   - Cursor foi modificado pelo cliente
	//   - Cursor está corrompido
	//   - Cursor não contém os campos esperados (sortField:sortValue:id)
	//
	// Como tratar:
	//   - Se o cursor veio do cliente: retornar erro 400 (Bad Request)
	//   - Se o cursor foi gerado pelo sistema: investigar bug no encoding
	//   - Sugerir ao cliente começar do início (sem cursor)
	ErrInvalidCursorFormat = errors.New("formato do cursor é inválido ou está corrompido")

	// ErrCursorExpired indica que o cursor não é mais válido.
	//
	// Regra de negócio violada: Cursores podem ter tempo de vida limitado
	// ou podem se tornar inválidos se os dados subjacentes mudarem drasticamente.
	//
	// Possíveis causas:
	//   - Cursor muito antigo (expirou)
	//   - Dados foram reorganizados (re-indexação)
	//   - Item referenciado pelo cursor foi deletado
	//
	// Como tratar: Solicitar ao cliente que comece uma nova navegação
	// do início (sem cursor).
	//
	// Nota: Este erro é opcional - nem todos os sistemas precisam implementar
	// expiração de cursor.
	ErrCursorExpired = errors.New("o cursor expirou ou não é mais válido")
)

// IsDomainError verifica se um erro é um erro de domínio de paginação.
//
// Útil para distinguir erros de domínio de erros técnicos (banco de dados,
// rede, etc.) e tratá-los de forma diferente.
//
// Exemplo de uso:
//
//	if cursor.IsDomainError(err) {
//	    // Retornar 400 Bad Request
//	    return BadRequestResponse(err.Error())
//	}
//	// Retornar 500 Internal Server Error
//	return InternalErrorResponse()
func IsDomainError(err error) bool {
	return errors.Is(err, ErrInvalidPageSize) ||
		errors.Is(err, ErrPageSizeTooLarge) ||
		errors.Is(err, ErrInvalidDirection) ||
		errors.Is(err, ErrInvalidCursorFormat) ||
		errors.Is(err, ErrCursorExpired)
}
