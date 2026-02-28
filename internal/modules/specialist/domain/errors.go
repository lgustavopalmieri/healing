package domain

import "errors"

var (
	ErrInvalidName          = errors.New("nome é obrigatório e deve ter pelo menos 2 caracteres")
	ErrInvalidEmail         = errors.New("email é obrigatório e deve ter formato válido")
	ErrInvalidSpecialty     = errors.New("especialidade é obrigatória")
	ErrInvalidLicenseNumber = errors.New("número da licença é obrigatório")
	ErrMustAgreeToShare     = errors.New("especialista deve concordar em compartilhar relatórios com pacientes")
	ErrInvalidStatus        = errors.New("status informado é inválido")
)
