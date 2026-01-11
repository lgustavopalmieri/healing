package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/utils"
)

func CreateSpecialist(
	name, email, phone, specialty, licenseNumber, description string,
	keywords []string,
	agreedToShare bool,
) (*Specialist, error) {
	if err := validate(name, email, specialty, licenseNumber, agreedToShare); err != nil {
		return nil, err
	}

	normalizedKeywords := utils.SanitizeStringArray(keywords)
	now := time.Now().UTC()

	specialist := &Specialist{
		ID:            uuid.New().String(),
		Name:          strings.TrimSpace(name),
		Email:         strings.ToLower(strings.TrimSpace(email)),
		Phone:         strings.TrimSpace(phone),
		Specialty:     strings.TrimSpace(specialty),
		LicenseNumber: strings.TrimSpace(licenseNumber),
		Description:   strings.TrimSpace(description),
		Keywords:      normalizedKeywords,
		AgreedToShare: agreedToShare,
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
