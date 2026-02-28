package update

import (
	"strings"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/utils"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

type UpdateSpecialistInput struct {
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

func UpdateSpecialist(existing *domain.Specialist, input UpdateSpecialistInput) (*domain.Specialist, error) {
	if err := validateID(input.ID); err != nil {
		return nil, err
	}

	if input.Name != nil {
		if err := domain.ValidateName(*input.Name); err != nil {
			return nil, err
		}
		existing.Name = strings.TrimSpace(*input.Name)
	}

	if input.Email != nil {
		if err := domain.ValidateEmail(*input.Email); err != nil {
			return nil, err
		}
		existing.Email = strings.ToLower(strings.TrimSpace(*input.Email))
	}

	if input.Phone != nil {
		existing.Phone = strings.TrimSpace(*input.Phone)
	}

	if input.Specialty != nil {
		if err := domain.ValidateSpecialty(*input.Specialty); err != nil {
			return nil, err
		}
		existing.Specialty = strings.TrimSpace(*input.Specialty)
	}

	if input.LicenseNumber != nil {
		if err := domain.ValidateLicenseNumber(*input.LicenseNumber); err != nil {
			return nil, err
		}
		existing.LicenseNumber = strings.TrimSpace(*input.LicenseNumber)
	}

	if input.Description != nil {
		existing.Description = strings.TrimSpace(*input.Description)
	}

	if input.Keywords != nil {
		existing.Keywords = utils.SanitizeStringArray(input.Keywords)
	}

	if input.AgreedToShare != nil {
		if err := domain.ValidateAgreedToShare(*input.AgreedToShare); err != nil {
			return nil, err
		}
		existing.AgreedToShare = *input.AgreedToShare
	}

	if input.Status != nil {
		if err := domain.ValidateStatus(*input.Status); err != nil {
			return nil, err
		}
		existing.Status = *input.Status
	}

	existing.UpdatedAt = time.Now().UTC()

	return existing, nil
}
