package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func strPtr(s string) *string                                      { return &s }
func boolPtr(b bool) *bool                                         { return &b }
func statusPtr(s domain.SpecialistStatus) *domain.SpecialistStatus { return &s }

func existingSpecialistFactory(overrides ...func(*domain.Specialist)) *domain.Specialist {
	now := time.Now().UTC()
	s := &domain.Specialist{
		ID:            "550e8400-e29b-41d4-a716-446655440000",
		Name:          "Dr. João Silva",
		Email:         "joao@example.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiologia",
		LicenseNumber: "CRM123456",
		Description:   "Especialista em cardiologia",
		Keywords:      []string{"coração", "arritmia"},
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

func updateDTOFactory(overrides ...func(*UpdateSpecialistDTO)) UpdateSpecialistDTO {
	dto := UpdateSpecialistDTO{
		ID:   "550e8400-e29b-41d4-a716-446655440000",
		Name: strPtr("Dr. Maria Santos"),
	}
	for _, o := range overrides {
		o(&dto)
	}
	return dto
}

func TestUpdateSpecialistUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		input          UpdateSpecialistDTO
		setupMocks     func(*mocks.MockSpecialistUpdateRepositoryInterface, *mocks.MockEventDispatcher)
		expectError    bool
		expectedErr    error
		validateResult func(*testing.T, *domain.Specialist)
	}{
		{
			name:  "success - updates specialist name and returns updated entity",
			input: updateDTOFactory(),
			setupMocks: func(mockRepo *mocks.MockSpecialistUpdateRepositoryInterface, mockEvent *mocks.MockEventDispatcher) {
				existing := existingSpecialistFactory()
				mockRepo.EXPECT().FindByID(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(existing, nil).Times(1)

				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, s *domain.Specialist) (*domain.Specialist, error) {
					return s, nil
				}).Times(1)

				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, s *domain.Specialist) {
				assert.NotNil(t, s)
				assert.Equal(t, "Dr. Maria Santos", s.Name)
				assert.Equal(t, "joao@example.com", s.Email)
				assert.Equal(t, domain.StatusActive, s.Status)
			},
		},
		{
			name: "success - updates multiple fields at once",
			input: updateDTOFactory(func(dto *UpdateSpecialistDTO) {
				dto.Name = strPtr("Dr. Carlos Souza")
				dto.Email = strPtr("carlos@example.com")
				dto.Specialty = strPtr("Neurologia")
				dto.Status = statusPtr(domain.StatusUnavailable)
			}),
			setupMocks: func(mockRepo *mocks.MockSpecialistUpdateRepositoryInterface, mockEvent *mocks.MockEventDispatcher) {
				existing := existingSpecialistFactory()
				mockRepo.EXPECT().FindByID(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(existing, nil).Times(1)

				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, s *domain.Specialist) (*domain.Specialist, error) {
					return s, nil
				}).Times(1)

				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, s *domain.Specialist) {
				assert.Equal(t, "Dr. Carlos Souza", s.Name)
				assert.Equal(t, "carlos@example.com", s.Email)
				assert.Equal(t, "Neurologia", s.Specialty)
				assert.Equal(t, domain.StatusUnavailable, s.Status)
			},
		},
		{
			name:  "success - still succeeds when event publish fails",
			input: updateDTOFactory(),
			setupMocks: func(mockRepo *mocks.MockSpecialistUpdateRepositoryInterface, mockEvent *mocks.MockEventDispatcher) {
				existing := existingSpecialistFactory()
				mockRepo.EXPECT().FindByID(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(existing, nil).Times(1)

				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, s *domain.Specialist) (*domain.Specialist, error) {
					return s, nil
				}).Times(1)

				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Return(errors.New("kafka unavailable")).Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, s *domain.Specialist) {
				assert.NotNil(t, s)
				assert.Equal(t, "Dr. Maria Santos", s.Name)
			},
		},
		{
			name:  "failure - returns ErrSpecialistNotFound when repository FindByID fails",
			input: updateDTOFactory(),
			setupMocks: func(mockRepo *mocks.MockSpecialistUpdateRepositoryInterface, mockEvent *mocks.MockEventDispatcher) {
				mockRepo.EXPECT().FindByID(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(nil, errors.New("not found")).Times(1)

				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrSpecialistNotFound,
		},
		{
			name: "failure - returns domain ErrInvalidName when name validation fails",
			input: updateDTOFactory(func(dto *UpdateSpecialistDTO) {
				dto.Name = strPtr("")
			}),
			setupMocks: func(mockRepo *mocks.MockSpecialistUpdateRepositoryInterface, mockEvent *mocks.MockEventDispatcher) {
				existing := existingSpecialistFactory()
				mockRepo.EXPECT().FindByID(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(existing, nil).Times(1)

				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: domain.ErrInvalidName,
		},
		{
			name: "failure - returns domain ErrInvalidEmail when email validation fails",
			input: updateDTOFactory(func(dto *UpdateSpecialistDTO) {
				dto.Name = nil
				dto.Email = strPtr("invalid-email")
			}),
			setupMocks: func(mockRepo *mocks.MockSpecialistUpdateRepositoryInterface, mockEvent *mocks.MockEventDispatcher) {
				existing := existingSpecialistFactory()
				mockRepo.EXPECT().FindByID(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(existing, nil).Times(1)

				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: domain.ErrInvalidEmail,
		},
		{
			name: "failure - returns domain ErrInvalidStatus when invalid status is provided",
			input: updateDTOFactory(func(dto *UpdateSpecialistDTO) {
				dto.Name = nil
				dto.Status = statusPtr(domain.SpecialistStatus("invalid"))
			}),
			setupMocks: func(mockRepo *mocks.MockSpecialistUpdateRepositoryInterface, mockEvent *mocks.MockEventDispatcher) {
				existing := existingSpecialistFactory()
				mockRepo.EXPECT().FindByID(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(existing, nil).Times(1)

				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: domain.ErrInvalidStatus,
		},
		{
			name:  "failure - returns ErrUpdateSpecialist when repository Update fails",
			input: updateDTOFactory(),
			setupMocks: func(mockRepo *mocks.MockSpecialistUpdateRepositoryInterface, mockEvent *mocks.MockEventDispatcher) {
				existing := existingSpecialistFactory()
				mockRepo.EXPECT().FindByID(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(existing, nil).Times(1)

				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("db connection lost")).Times(1)

				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrUpdateSpecialist,
		},
		{
			name: "failure - returns domain ErrMustAgreeToShare when agreed_to_share is set to false",
			input: updateDTOFactory(func(dto *UpdateSpecialistDTO) {
				dto.Name = nil
				dto.AgreedToShare = boolPtr(false)
			}),
			setupMocks: func(mockRepo *mocks.MockSpecialistUpdateRepositoryInterface, mockEvent *mocks.MockEventDispatcher) {
				existing := existingSpecialistFactory()
				mockRepo.EXPECT().FindByID(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").Return(existing, nil).Times(1)

				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: domain.ErrMustAgreeToShare,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSpecialistUpdateRepositoryInterface(ctrl)
			mockEvent := mocks.NewMockEventDispatcher(ctrl)

			tt.setupMocks(mockRepo, mockEvent)

			useCase := NewUpdateSpecialistUseCase(mockRepo, mockEvent)

			result, err := useCase.Execute(context.Background(), tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}
