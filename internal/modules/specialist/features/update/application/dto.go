package application

import "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"

type UpdateSpecialistDTO struct {
	ID            string
	Name          *string
	Email         *string
	Phone         *string
	Specialty     *string
	LicenseNumber *string
	Description   *string
	Keywords      []string
	AgreedToShare *bool
	Status        *domain.SpecialistStatus
}
