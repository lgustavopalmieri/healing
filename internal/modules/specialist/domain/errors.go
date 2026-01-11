package domain

import "errors"

// Domain errors
var (
	ErrInvalidName          = errors.New("nome é obrigatório e deve ter pelo menos 2 caracteres")
	ErrInvalidEmail         = errors.New("email é obrigatório e deve ter formato válido")
	ErrInvalidSpecialty     = errors.New("especialidade é obrigatória")
	ErrInvalidLicenseNumber = errors.New("número da licença é obrigatório")
	ErrMustAgreeToShare     = errors.New("especialista deve concordar em compartilhar relatórios com pacientes")
	ErrDuplicateEmail       = errors.New("email já está em uso por outro especialista")
	ErrDuplicateLicense     = errors.New("número da licença já está em uso por outro especialista")
	ErrDuplicateID          = errors.New("ID já existe no sistema")
	ErrInvalidLicense       = errors.New("número da licença é inválido ou não foi possível validar")
)
