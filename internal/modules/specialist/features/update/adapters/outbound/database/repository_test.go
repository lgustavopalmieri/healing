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
	createdb "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/adapters/outbound/database"
)

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
	s := &domain.Specialist{
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
	for _, o := range overrides {
		o(s)
	}
	return s
}

func seedSpecialist(t *testing.T, db *sql.DB, s *domain.Specialist) {
	createRepo := createdb.NewSpecialistCreateRepository(db)
	_, err := createRepo.SaveWithValidation(context.Background(), s)
	require.NoError(t, err)
}

func TestSpecialistUpdateRepository_FindByID(t *testing.T) {
	tests := []struct {
		name           string
		setupInput     func(*sql.DB) string
		expectError    bool
		expectedErrMsg string
		validateResult func(*testing.T, *domain.Specialist)
	}{
		{
			name: "success - finds specialist by ID and returns all fields",
			setupInput: func(db *sql.DB) string {
				s := specialistFactory()
				seedSpecialist(t, db, s)
				return s.ID
			},
			expectError: false,
			validateResult: func(t *testing.T, result *domain.Specialist) {
				assert.NotNil(t, result)
				assert.Equal(t, "Dr. João Silva", result.Name)
				assert.Equal(t, "Cardiologia", result.Specialty)
				assert.Equal(t, domain.StatusActive, result.Status)
				assert.NotEmpty(t, result.ID)
			},
		},
		{
			name: "failure - returns error when specialist does not exist",
			setupInput: func(db *sql.DB) string {
				return uuid.New().String()
			},
			expectError:    true,
			expectedErrMsg: "not found",
			validateResult: func(t *testing.T, result *domain.Specialist) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup := setupTestDB(t)
			defer cleanup()

			repo := NewSpecialistUpdateRepository(db)
			id := tt.setupInput(db)

			result, err := repo.FindByID(context.Background(), id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, result)
		})
	}
}

func TestSpecialistUpdateRepository_Update(t *testing.T) {
	tests := []struct {
		name           string
		setupInput     func(*sql.DB) *domain.Specialist
		expectError    bool
		expectedErrMsg string
		validateResult func(*testing.T, *domain.Specialist, *domain.Specialist)
	}{
		{
			name: "success - updates specialist and returns updated data with all fields",
			setupInput: func(db *sql.DB) *domain.Specialist {
				s := specialistFactory()
				seedSpecialist(t, db, s)
				s.Name = "Dr. Maria Santos"
				s.Email = "maria@example.com"
				s.Specialty = "Neurologia"
				s.Status = domain.StatusUnavailable
				s.UpdatedAt = time.Now().UTC()
				return s
			},
			expectError: false,
			validateResult: func(t *testing.T, input, result *domain.Specialist) {
				assert.NotNil(t, result)
				assert.Equal(t, input.ID, result.ID)
				assert.Equal(t, "Dr. Maria Santos", result.Name)
				assert.Equal(t, "maria@example.com", result.Email)
				assert.Equal(t, "Neurologia", result.Specialty)
				assert.Equal(t, domain.StatusUnavailable, result.Status)
				assert.Equal(t, input.AgreedToShare, result.AgreedToShare)
				assert.Equal(t, input.Keywords, result.Keywords)
			},
		},
		{
			name: "success - updates only changed fields preserving others",
			setupInput: func(db *sql.DB) *domain.Specialist {
				s := specialistFactory()
				seedSpecialist(t, db, s)
				s.Description = "Nova descrição atualizada"
				s.UpdatedAt = time.Now().UTC()
				return s
			},
			expectError: false,
			validateResult: func(t *testing.T, input, result *domain.Specialist) {
				assert.NotNil(t, result)
				assert.Equal(t, "Nova descrição atualizada", result.Description)
				assert.Equal(t, input.Name, result.Name)
				assert.Equal(t, input.Email, result.Email)
				assert.Equal(t, input.Specialty, result.Specialty)
				assert.Equal(t, input.LicenseNumber, result.LicenseNumber)
			},
		},
		{
			name: "failure - returns error when specialist ID does not exist",
			setupInput: func(db *sql.DB) *domain.Specialist {
				return specialistFactory(func(s *domain.Specialist) {
					s.ID = uuid.New().String()
				})
			},
			expectError:    true,
			expectedErrMsg: "not found",
			validateResult: func(t *testing.T, input, result *domain.Specialist) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup := setupTestDB(t)
			defer cleanup()

			repo := NewSpecialistUpdateRepository(db)
			input := tt.setupInput(db)

			result, err := repo.Update(context.Background(), input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, input, result)
		})
	}
}
