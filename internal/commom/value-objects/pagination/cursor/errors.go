package cursor

import "errors"

var (
	ErrInvalidPageSize     = errors.New("o tamanho da página deve ser maior que zero")
	ErrPageSizeTooLarge    = errors.New("o tamanho da página excede o limite máximo permitido (100 itens)")
	ErrInvalidDirection    = errors.New("direção de paginação inválida: use 'next' ou 'previous'")
	ErrInvalidCursorFormat = errors.New("formato do cursor é inválido ou está corrompido")
	ErrCursorExpired       = errors.New("o cursor expirou ou não é mais válido")
)

func IsDomainError(err error) bool {
	return errors.Is(err, ErrInvalidPageSize) ||
		errors.Is(err, ErrPageSizeTooLarge) ||
		errors.Is(err, ErrInvalidDirection) ||
		errors.Is(err, ErrInvalidCursorFormat) ||
		errors.Is(err, ErrCursorExpired)
}
