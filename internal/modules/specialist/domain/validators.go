package domain

import (
	"regexp"
	"strings"
)

func validateName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return ErrInvalidName
	}
	return nil
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrInvalidEmail
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

func validateSpecialty(specialty string) error {
	specialty = strings.TrimSpace(specialty)
	if specialty == "" {
		return ErrInvalidSpecialty
	}
	return nil
}

func validateLicenseNumber(licenseNumber string) error {
	licenseNumber = strings.TrimSpace(licenseNumber)
	if licenseNumber == "" {
		return ErrInvalidLicenseNumber
	}
	return nil
}
