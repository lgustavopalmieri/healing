package domain

import (
	"regexp"
	"strings"
)

func ValidateName(name string) error {
	if len(strings.TrimSpace(name)) < 2 {
		return ErrInvalidName
	}
	return nil
}

func ValidateEmail(email string) error {
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

func ValidateSpecialty(specialty string) error {
	if strings.TrimSpace(specialty) == "" {
		return ErrInvalidSpecialty
	}
	return nil
}

func ValidateLicenseNumber(licenseNumber string) error {
	if strings.TrimSpace(licenseNumber) == "" {
		return ErrInvalidLicenseNumber
	}
	return nil
}

func ValidateAgreedToShare(agreed bool) error {
	if !agreed {
		return ErrMustAgreeToShare
	}
	return nil
}

func ValidateStatus(status SpecialistStatus) error {
	switch status {
	case StatusPending, StatusActive, StatusUnavailable, StatusDeleted, StatusBanned:
		return nil
	default:
		return ErrInvalidStatus
	}
}
