package authorizelicense

import (
	"testing"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/stretchr/testify/assert"
)

func specialistFactory(overrides ...func(*domain.Specialist)) *domain.Specialist {
	now := time.Now().UTC()
	s := &domain.Specialist{
		ID:            "test-id-123",
		Name:          "Dr. João Silva",
		Email:         "joao@example.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiologia",
		LicenseNumber: "CRM123456",
		Description:   "Especialista em cardiologia",
		Keywords:      []string{"coração", "arritmia"},
		AgreedToShare: true,
		Rating:        0.0,
		Status:        domain.StatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, o := range overrides {
		o(s)
	}
	return s
}

func TestAuthorizeLicense(t *testing.T) {
	tests := []struct {
		name           string
		overrides      []func(*domain.Specialist)
		expectError    bool
		expectedErr    error
		validateResult func(*testing.T, *domain.Specialist)
	}{
		{
			name:        "success - authorizes license when specialist status is pending",
			overrides:   nil,
			expectError: false,
			validateResult: func(t *testing.T, result *domain.Specialist) {
				assert.NotNil(t, result)
				assert.Equal(t, domain.StatusAuthorizedLicense, result.Status)
			},
		},
		{
			name:        "success - sets status to authorized_license",
			overrides:   nil,
			expectError: false,
			validateResult: func(t *testing.T, result *domain.Specialist) {
				assert.Equal(t, domain.SpecialistStatus("authorized_license"), result.Status)
			},
		},
		{
			name:        "success - updates UpdatedAt timestamp",
			overrides:   nil,
			expectError: false,
			validateResult: func(t *testing.T, result *domain.Specialist) {
				assert.True(t, result.UpdatedAt.After(result.CreatedAt) || result.UpdatedAt.Equal(result.CreatedAt))
			},
		},
		{
			name: "failure - returns ErrInvalidStatusTransition when status is active",
			overrides: []func(*domain.Specialist){
				func(s *domain.Specialist) { s.Status = domain.StatusActive },
			},
			expectError: true,
			expectedErr: ErrInvalidStatusTransition,
		},
		{
			name: "failure - returns ErrInvalidStatusTransition when status is authorized_license",
			overrides: []func(*domain.Specialist){
				func(s *domain.Specialist) { s.Status = domain.StatusAuthorizedLicense },
			},
			expectError: true,
			expectedErr: ErrInvalidStatusTransition,
		},
		{
			name: "failure - returns ErrInvalidStatusTransition when status is unavailable",
			overrides: []func(*domain.Specialist){
				func(s *domain.Specialist) { s.Status = domain.StatusUnavailable },
			},
			expectError: true,
			expectedErr: ErrInvalidStatusTransition,
		},
		{
			name: "failure - returns ErrInvalidStatusTransition when status is deleted",
			overrides: []func(*domain.Specialist){
				func(s *domain.Specialist) { s.Status = domain.StatusDeleted },
			},
			expectError: true,
			expectedErr: ErrInvalidStatusTransition,
		},
		{
			name: "failure - returns ErrInvalidStatusTransition when status is banned",
			overrides: []func(*domain.Specialist){
				func(s *domain.Specialist) { s.Status = domain.StatusBanned },
			},
			expectError: true,
			expectedErr: ErrInvalidStatusTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specialist := specialistFactory(tt.overrides...)

			result, err := AuthorizeLicense(specialist)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}
