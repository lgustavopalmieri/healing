package create

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

var (
	ErrInvalidName          = domain.ErrInvalidName
	ErrInvalidEmail         = domain.ErrInvalidEmail
	ErrInvalidSpecialty     = domain.ErrInvalidSpecialty
	ErrInvalidLicenseNumber = domain.ErrInvalidLicenseNumber
	ErrMustAgreeToShare     = domain.ErrMustAgreeToShare
)
