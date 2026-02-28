package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type validatorTestCase struct {
	name        string
	input       string
	expectError bool
	expectedErr error
}

func TestValidateName(t *testing.T) {
	tests := []validatorTestCase{
		{
			name:        "should return nil when name is valid",
			input:       "João Silva",
			expectError: false,
		},
		{
			name:        "should return nil when name has exactly 2 characters",
			input:       "Jo",
			expectError: false,
		},
		{
			name:        "should return error when name is empty",
			input:       "",
			expectError: true,
			expectedErr: ErrInvalidName,
		},
		{
			name:        "should return error when name has only 1 character",
			input:       "J",
			expectError: true,
			expectedErr: ErrInvalidName,
		},
		{
			name:        "should return error when name is only whitespace",
			input:       "   ",
			expectError: true,
			expectedErr: ErrInvalidName,
		},
		{
			name:        "should trim whitespace and validate correctly",
			input:       "  João  ",
			expectError: false,
		},
		{
			name:        "should return error when trimmed name becomes too short",
			input:       "  J  ",
			expectError: true,
			expectedErr: ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []validatorTestCase{
		{
			name:        "should return nil when email is valid",
			input:       "joao@example.com",
			expectError: false,
		},
		{
			name:        "should return nil when email has numbers and special characters",
			input:       "user123+test@domain.co.uk",
			expectError: false,
		},
		{
			name:        "should return nil when email has dots and underscores",
			input:       "first.last_name@company.org",
			expectError: false,
		},
		{
			name:        "should return error when email is empty",
			input:       "",
			expectError: true,
			expectedErr: ErrInvalidEmail,
		},
		{
			name:        "should return error when email is only whitespace",
			input:       "   ",
			expectError: true,
			expectedErr: ErrInvalidEmail,
		},
		{
			name:        "should return error when email has no @ symbol",
			input:       "invalid-email.com",
			expectError: true,
			expectedErr: ErrInvalidEmail,
		},
		{
			name:        "should return error when email has no domain",
			input:       "user@",
			expectError: true,
			expectedErr: ErrInvalidEmail,
		},
		{
			name:        "should return error when email has no local part",
			input:       "@domain.com",
			expectError: true,
			expectedErr: ErrInvalidEmail,
		},
		{
			name:        "should return error when email has invalid domain extension",
			input:       "user@domain.c",
			expectError: true,
			expectedErr: ErrInvalidEmail,
		},
		{
			name:        "should return error when email has multiple @ symbols",
			input:       "user@@domain.com",
			expectError: true,
			expectedErr: ErrInvalidEmail,
		},
		{
			name:        "should trim whitespace and validate correctly",
			input:       "  user@domain.com  ",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSpecialty(t *testing.T) {
	tests := []validatorTestCase{
		{
			name:        "should return nil when specialty is valid",
			input:       "Cardiologia",
			expectError: false,
		},
		{
			name:        "should return nil when specialty has special characters",
			input:       "Ortopedia & Traumatologia",
			expectError: false,
		},
		{
			name:        "should return error when specialty is empty",
			input:       "",
			expectError: true,
			expectedErr: ErrInvalidSpecialty,
		},
		{
			name:        "should return error when specialty is only whitespace",
			input:       "   ",
			expectError: true,
			expectedErr: ErrInvalidSpecialty,
		},
		{
			name:        "should trim whitespace and validate correctly",
			input:       "  Neurologia  ",
			expectError: false,
		},
		{
			name:        "should return error when trimmed specialty becomes empty",
			input:       "  \t\n  ",
			expectError: true,
			expectedErr: ErrInvalidSpecialty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpecialty(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLicenseNumber(t *testing.T) {
	tests := []validatorTestCase{
		{
			name:        "should return nil when license number is valid",
			input:       "CRM123456",
			expectError: false,
		},
		{
			name:        "should return nil when license number has numbers only",
			input:       "123456",
			expectError: false,
		},
		{
			name:        "should return nil when license number has mixed format",
			input:       "CRM-123456/SP",
			expectError: false,
		},
		{
			name:        "should return error when license number is empty",
			input:       "",
			expectError: true,
			expectedErr: ErrInvalidLicenseNumber,
		},
		{
			name:        "should return error when license number is only whitespace",
			input:       "   ",
			expectError: true,
			expectedErr: ErrInvalidLicenseNumber,
		},
		{
			name:        "should trim whitespace and validate correctly",
			input:       "  CRM123456  ",
			expectError: false,
		},
		{
			name:        "should return error when trimmed license number becomes empty",
			input:       "  \t\n  ",
			expectError: true,
			expectedErr: ErrInvalidLicenseNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLicenseNumber(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAgreedToShare(t *testing.T) {
	tests := []struct {
		name        string
		input       bool
		expectError bool
		expectedErr error
	}{
		{
			name:        "should return nil when agreed is true",
			input:       true,
			expectError: false,
		},
		{
			name:        "should return error when agreed is false",
			input:       false,
			expectError: true,
			expectedErr: ErrMustAgreeToShare,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgreedToShare(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name        string
		input       SpecialistStatus
		expectError bool
		expectedErr error
	}{
		{name: "should return nil for pending", input: StatusPending, expectError: false},
		{name: "should return nil for active", input: StatusActive, expectError: false},
		{name: "should return nil for unavailable", input: StatusUnavailable, expectError: false},
		{name: "should return nil for deleted", input: StatusDeleted, expectError: false},
		{name: "should return nil for banned", input: StatusBanned, expectError: false},
		{name: "should return error for invalid status", input: SpecialistStatus("invalid"), expectError: true, expectedErr: ErrInvalidStatus},
		{name: "should return error for empty status", input: SpecialistStatus(""), expectError: true, expectedErr: ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatus(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
