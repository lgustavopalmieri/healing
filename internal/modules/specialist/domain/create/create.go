package create

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/utils"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

type CreateSpecialistInput struct {
	Name          string
	Email         string
	Phone         string
	Specialty     string
	LicenseNumber string
	Description   string
	Keywords      []string
	AgreedToShare bool
}

func CreateSpecialist(input CreateSpecialistInput) (*domain.Specialist, error) {
	if err := validate(input.Name, input.Email, input.Specialty, input.LicenseNumber, input.AgreedToShare); err != nil {
		return nil, err
	}

	normalizedKeywords := utils.SanitizeStringArray(input.Keywords)
	now := time.Now().UTC()

	specialist := &domain.Specialist{
		ID:            uuid.New().String(),
		Name:          strings.TrimSpace(input.Name),
		Email:         strings.ToLower(strings.TrimSpace(input.Email)),
		Phone:         strings.TrimSpace(input.Phone),
		Specialty:     strings.TrimSpace(input.Specialty),
		LicenseNumber: strings.TrimSpace(input.LicenseNumber),
		Description:   strings.TrimSpace(input.Description),
		Keywords:      normalizedKeywords,
		AgreedToShare: input.AgreedToShare,
		Rating:        0.0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return specialist, nil
}

func validate(name, email, specialty, licenseNumber string, agreedToShare bool) error {
	if err := validateName(name); err != nil {
		return err
	}

	if err := validateEmail(email); err != nil {
		return err
	}

	if err := validateSpecialty(specialty); err != nil {
		return err
	}

	if err := validateLicenseNumber(licenseNumber); err != nil {
		return err
	}

	if !agreedToShare {
		return ErrMustAgreeToShare
	}

	return nil
}
