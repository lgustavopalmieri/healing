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

func seedSpecialist(t *testing.T, db *sql.DB, s *domain.Specialist) {
	createRepo := createdb.NewSpecialistCreateRepository(db)
	_, err := createRepo.SaveWithValidation(context.Background(), s)
	require.NoError(t, err)
}

func TestValidateLicenseRepository_FindByID(t *testing.T) {
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
				assert.Equal(t, domain.StatusPending, result.Status)
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

			repo := NewValidateLicenseRepository(db)
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

func TestValidateLicenseRepository_UpdateStatus(t *testing.T) {
	tests := []struct {
		name           string
		setupInput     func(*sql.DB) (string, domain.SpecialistStatus)
		expectError    bool
		expectedErrMsg string
		validate       func(*testing.T, *sql.DB, string)
	}{
		{
			name: "success - updates only status and updated_at fields",
			setupInput: func(db *sql.DB) (string, domain.SpecialistStatus) {
				s := specialistFactory()
				seedSpecialist(t, db, s)
				return s.ID, domain.StatusAuthorizedLicense
			},
			expectError: false,
			validate: func(t *testing.T, db *sql.DB, id string) {
				repo := NewValidateLicenseRepository(db)
				result, err := repo.FindByID(context.Background(), id)
				require.NoError(t, err)
				assert.Equal(t, domain.StatusAuthorizedLicense, result.Status)
			},
		},
		{
			name: "success - preserves all other specialist fields unchanged",
			setupInput: func(db *sql.DB) (string, domain.SpecialistStatus) {
				s := specialistFactory()
				seedSpecialist(t, db, s)
				return s.ID, domain.StatusAuthorizedLicense
			},
			expectError: false,
			validate: func(t *testing.T, db *sql.DB, id string) {
				repo := NewValidateLicenseRepository(db)
				result, err := repo.FindByID(context.Background(), id)
				require.NoError(t, err)
				assert.Equal(t, "Dr. João Silva", result.Name)
				assert.Equal(t, "Cardiologia", result.Specialty)
				assert.True(t, result.AgreedToShare)
				assert.Equal(t, 0.0, result.Rating)
				assert.Equal(t, domain.StatusAuthorizedLicense, result.Status)
			},
		},
		{
			name: "failure - returns error when specialist ID does not exist",
			setupInput: func(db *sql.DB) (string, domain.SpecialistStatus) {
				return uuid.New().String(), domain.StatusAuthorizedLicense
			},
			expectError:    true,
			expectedErrMsg: "not found",
			validate:       func(t *testing.T, db *sql.DB, id string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup := setupTestDB(t)
			defer cleanup()

			repo := NewValidateLicenseRepository(db)
			id, status := tt.setupInput(db)

			err := repo.UpdateStatus(context.Background(), id, status)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				tt.validate(t, db, id)
			}
		})
	}
}
