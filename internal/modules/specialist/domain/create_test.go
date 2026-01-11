package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createSpecialistFactory(overrides ...func(*Specialist)) (*Specialist, error) {
	specialist, err := CreateSpecialist(
		"Dr. João Silva",
		"joao@example.com",
		"+5511999999999",
		"Cardiologia",
		"CRM123456",
		"Especialista em cardiologia",
		[]string{"coração", "arritmia"},
		true,
	)
	if err != nil {
		return nil, err
	}

	for _, override := range overrides {
		override(specialist)
	}

	return specialist, nil
}

func TestCreateSpecialist(t *testing.T) {
	tests := []struct {
		name           string
		overrides      []func(*Specialist)
		expectError    bool
		expectedErr    error
		validateResult func(*testing.T, *Specialist)
	}{
		{
			name:        "should create specialist successfully with valid data",
			overrides:   nil,
			expectError: false,
			validateResult: func(t *testing.T, specialist *Specialist) {
				assert.NotEmpty(t, specialist.ID)
				assert.Equal(t, "Dr. João Silva", specialist.Name)
				assert.Equal(t, "joao@example.com", specialist.Email)
				assert.Equal(t, "+5511999999999", specialist.Phone)
				assert.Equal(t, "Cardiologia", specialist.Specialty)
				assert.Equal(t, "CRM123456", specialist.LicenseNumber)
				assert.Equal(t, "Especialista em cardiologia", specialist.Description)
				assert.Equal(t, []string{"coração", "arritmia"}, specialist.Keywords)
				assert.True(t, specialist.AgreedToShare)
				assert.False(t, specialist.CreatedAt.IsZero())
				assert.False(t, specialist.UpdatedAt.IsZero())
				assert.Equal(t, specialist.CreatedAt, specialist.UpdatedAt)
			},
		},
		{
			name: "should return error when name is invalid",
			overrides: []func(*Specialist){
				func(s *Specialist) { s.Name = "J" },
			},
			expectError: true,
			expectedErr: ErrInvalidName,
		},
		{
			name: "should return error when email is invalid",
			overrides: []func(*Specialist){
				func(s *Specialist) { s.Email = "invalid-email" },
			},
			expectError: true,
			expectedErr: ErrInvalidEmail,
		},
		{
			name: "should return error when specialty is invalid",
			overrides: []func(*Specialist){
				func(s *Specialist) { s.Specialty = "" },
			},
			expectError: true,
			expectedErr: ErrInvalidSpecialty,
		},
		{
			name: "should return error when license number is invalid",
			overrides: []func(*Specialist){
				func(s *Specialist) { s.LicenseNumber = "" },
			},
			expectError: true,
			expectedErr: ErrInvalidLicenseNumber,
		},
		{
			name: "should return error when agreed to share is false",
			overrides: []func(*Specialist){
				func(s *Specialist) { s.AgreedToShare = false },
			},
			expectError: true,
			expectedErr: ErrMustAgreeToShare,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specialist, err := createSpecialistFactory(tt.overrides...)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, specialist)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, specialist)
				if tt.validateResult != nil {
					tt.validateResult(t, specialist)
				}
			}
		})
	}
}

func TestSanitizesWhenCreate(t *testing.T) {
	t.Run("should trim whitespace from all string fields", func(t *testing.T) {
		specialist, err := CreateSpecialist(
			"  Dr. João Silva  ",
			"  JOAO@EXAMPLE.COM  ",
			"  +5511999999999  ",
			"  Cardiologia  ",
			"  CRM123456  ",
			"  Especialista em cardiologia  ",
			[]string{"  coração  ", "arritmia"},
			true,
		)

		assert.NoError(t, err)
		assert.Equal(t, "Dr. João Silva", specialist.Name)
		assert.Equal(t, "joao@example.com", specialist.Email)
		assert.Equal(t, "+5511999999999", specialist.Phone)
		assert.Equal(t, "Cardiologia", specialist.Specialty)
		assert.Equal(t, "CRM123456", specialist.LicenseNumber)
		assert.Equal(t, "Especialista em cardiologia", specialist.Description)
		assert.Equal(t, []string{"coração", "arritmia"}, specialist.Keywords)
	})

	t.Run("should normalize keywords using sanitize function", func(t *testing.T) {
		specialist, err := CreateSpecialist(
			"Dr. João",
			"joao@example.com",
			"",
			"Cardiologia",
			"CRM123",
			"",
			[]string{"CORAÇÃO", "coração", "  Arritmia  ", "", "hipertensão", "HIPERTENSÃO"},
			true,
		)

		assert.NoError(t, err)
		assert.Equal(t, []string{"coração", "arritmia", "hipertensão"}, specialist.Keywords)
	})

	t.Run("should handle empty keywords array", func(t *testing.T) {
		specialist, err := CreateSpecialist("Dr. João", "joao@example.com", "", "Cardiologia", "CRM123", "", []string{}, true)

		assert.NoError(t, err)
		assert.Equal(t, []string{}, specialist.Keywords)
	})

	t.Run("should handle nil keywords array", func(t *testing.T) {
		specialist, err := CreateSpecialist("Dr. João", "joao@example.com", "", "Cardiologia", "CRM123", "", nil, true)

		assert.NoError(t, err)
		assert.Equal(t, []string{}, specialist.Keywords)
	})

	t.Run("should generate unique IDs for different specialists", func(t *testing.T) {
		specialist1, err1 := createSpecialistFactory()
		specialist2, err2 := createSpecialistFactory(
			func(s *Specialist) {
				s.Name = "Dr. Maria"
				s.Email = "maria@example.com"
				s.Specialty = "Neurologia"
				s.LicenseNumber = "CRM456"
			},
		)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, specialist1.ID, specialist2.ID)
	})

	t.Run("should set timestamps close to current time", func(t *testing.T) {
		before := time.Now().UTC()
		specialist, err := createSpecialistFactory()
		after := time.Now().UTC()

		assert.NoError(t, err)
		assert.True(t, specialist.CreatedAt.After(before) || specialist.CreatedAt.Equal(before))
		assert.True(t, specialist.CreatedAt.Before(after) || specialist.CreatedAt.Equal(after))
		assert.True(t, specialist.UpdatedAt.After(before) || specialist.UpdatedAt.Equal(before))
		assert.True(t, specialist.UpdatedAt.Before(after) || specialist.UpdatedAt.Equal(after))
	})
}
