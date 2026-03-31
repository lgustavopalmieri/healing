package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/database/postgresql"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
)

// go test ./internal/modules/specialist/features/create/adapters/outbound/database/... -v

var testHelper = postgresql.NewTestHelper()

func TestMain(m *testing.M) {
	testHelper.RunTestMain(m)
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	return testHelper.SetupTestDB(t)
}

func specialistFactory(overrides ...func(*domain.Specialist)) *domain.Specialist {
	now := time.Now().UTC()
	uniqueID := uuid.New().String()
	specialist := &domain.Specialist{
		ID:            uniqueID,
		Name:          "Dr. João Silva",
		Email:         "joao.silva+" + uniqueID[:8] + "@example.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiologia",
		LicenseNumber: "CRM" + uniqueID[:6],
		Description:   "Cardiologista especializado em arritmias",
		Keywords:      []string{"cardiologia", "arritmia", "coração"},
		AgreedToShare: true,
		Rating:        4.5,
		Status:        domain.StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	for _, override := range overrides {
		override(specialist)
	}

	return specialist
}

func TestSpecialistCreateRepository_SaveWithValidation(t *testing.T) {
	tests := []struct {
		name           string
		setupInput     func(application.SpecialistCreateRepositoryInterface) *domain.Specialist
		expectError    bool
		expectedErr    string
		validateResult func(*testing.T, *domain.Specialist, *domain.Specialist)
	}{
		{
			name: "success - validates uniqueness and saves specialist atomically",
			setupInput: func(repo application.SpecialistCreateRepositoryInterface) *domain.Specialist {
				return specialistFactory()
			},
			expectError: false,
			validateResult: func(t *testing.T, input, result *domain.Specialist) {
				assert.Equal(t, input.ID, result.ID)
				assert.Equal(t, input.Name, result.Name)
				assert.Equal(t, input.Email, result.Email)
				assert.Equal(t, input.Phone, result.Phone)
				assert.Equal(t, input.Specialty, result.Specialty)
				assert.Equal(t, input.LicenseNumber, result.LicenseNumber)
				assert.Equal(t, input.Description, result.Description)
				assert.Equal(t, input.Keywords, result.Keywords)
				assert.Equal(t, input.AgreedToShare, result.AgreedToShare)
				assert.WithinDuration(t, input.CreatedAt, result.CreatedAt, time.Second)
				assert.WithinDuration(t, input.UpdatedAt, result.UpdatedAt, time.Second)
			},
		},
		{
			name: "failure - returns error when ID already exists",
			setupInput: func(repo application.SpecialistCreateRepositoryInterface) *domain.Specialist {
				existing := specialistFactory()
				_, err := repo.SaveWithValidation(context.Background(), existing)
				require.NoError(t, err)

				return specialistFactory(func(s *domain.Specialist) {
					s.ID = existing.ID
				})
			},
			expectError: true,
			expectedErr: "already exists",
			validateResult: func(t *testing.T, input, result *domain.Specialist) {
				assert.Nil(t, result)
			},
		},
		{
			name: "failure - returns error when email already exists",
			setupInput: func(repo application.SpecialistCreateRepositoryInterface) *domain.Specialist {
				existing := specialistFactory(func(s *domain.Specialist) {
					s.Email = "existing@example.com"
				})
				_, err := repo.SaveWithValidation(context.Background(), existing)
				require.NoError(t, err)

				return specialistFactory(func(s *domain.Specialist) {
					s.Email = "existing@example.com"
				})
			},
			expectError: true,
			expectedErr: "already exists",
			validateResult: func(t *testing.T, input, result *domain.Specialist) {
				assert.Nil(t, result)
			},
		},
		{
			name: "failure - returns error when license number already exists",
			setupInput: func(repo application.SpecialistCreateRepositoryInterface) *domain.Specialist {
				existing := specialistFactory(func(s *domain.Specialist) {
					s.LicenseNumber = "CRM555555"
				})
				_, err := repo.SaveWithValidation(context.Background(), existing)
				require.NoError(t, err)

				return specialistFactory(func(s *domain.Specialist) {
					s.LicenseNumber = "CRM555555"
				})
			},
			expectError: true,
			expectedErr: "already exists",
			validateResult: func(t *testing.T, input, result *domain.Specialist) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup := setupTestDB(t)
			defer cleanup()

			repo := NewSpecialistCreateRepository(db)
			input := tt.setupInput(repo)

			result, err := repo.SaveWithValidation(context.Background(), input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			if tt.validateResult != nil {
				tt.validateResult(t, input, result)
			}
		})
	}
}
