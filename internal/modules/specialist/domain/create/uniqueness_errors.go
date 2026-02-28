package create

import "errors"

var (
	ErrDuplicateEmail   = errors.New("email já está em uso por outro especialista")
	ErrDuplicateLicense = errors.New("número da licença já está em uso por outro especialista")
	ErrDuplicateID      = errors.New("ID já existe no sistema")
	ErrInvalidLicense   = errors.New("número da licença é inválido ou não foi possível validar")
)
