package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/create"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application/mocks"

	loggermocks "github.com/lgustavopalmieri/healing-specialist/internal/commom/observability/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func createSpecialistFactory(overrides ...func(*CreateSpecialistDTO)) *CreateSpecialistDTO {
	input := &CreateSpecialistDTO{
		Name:          "Dr. João Silva",
		Email:         "joao@example.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiologia",
		LicenseNumber: "CRM123456",
		Description:   "Especialista em cardiologia",
		Keywords:      []string{"coração", "arritmia"},
		AgreedToShare: true,
	}
	for _, override := range overrides {
		override(input)
	}
	return input
}

func TestCreateSpecialistUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		inputOverrides []func(*CreateSpecialistDTO)
		setupMocks     func(*mocks.MockSpecialistCreateRepositoryInterface, *mocks.MockEventDispatcher, *loggermocks.MockLogger, CreateSpecialistDTO) chan struct{}
		expectError    bool
		expectedErr    error
		validateResult func(*testing.T, *domain.Specialist)
	}{
		{
			name:           "creates specialist with status pending and publishes event asynchronously",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockEventPublisher *mocks.MockEventDispatcher, mockLogger *loggermocks.MockLogger, input CreateSpecialistDTO) chan struct{} {
				done := make(chan struct{})

				mockRepo.EXPECT().SaveWithValidation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error) {
					return specialist, nil
				}).Times(1)

				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, evt interface{}) error {
					defer close(done)
					return nil
				}).Times(1)

				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				return done
			},
			expectError: false,
			validateResult: func(t *testing.T, specialist *domain.Specialist) {
				assert.NotNil(t, specialist)
				assert.Equal(t, "Dr. João Silva", specialist.Name)
				assert.Equal(t, "joao@example.com", specialist.Email)
				assert.Equal(t, "CRM123456", specialist.LicenseNumber)
				assert.Equal(t, "Cardiologia", specialist.Specialty)
				assert.Equal(t, "+5511999999999", specialist.Phone)
				assert.Equal(t, "Especialista em cardiologia", specialist.Description)
				assert.Equal(t, []string{"coração", "arritmia"}, specialist.Keywords)
				assert.True(t, specialist.AgreedToShare)
				assert.Equal(t, 0.0, specialist.Rating)
				assert.Equal(t, domain.StatusPending, specialist.Status)
				assert.NotEmpty(t, specialist.ID)
				assert.False(t, specialist.CreatedAt.IsZero())
				assert.False(t, specialist.UpdatedAt.IsZero())
			},
		},
		{
			name: "returns domain error when name is invalid",
			inputOverrides: []func(*CreateSpecialistDTO){
				func(dto *CreateSpecialistDTO) { dto.Name = "" },
			},
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockEventPublisher *mocks.MockEventDispatcher, mockLogger *loggermocks.MockLogger, input CreateSpecialistDTO) chan struct{} {
				mockRepo.EXPECT().SaveWithValidation(gomock.Any(), gomock.Any()).Times(0)
				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
				return nil
			},
			expectError: true,
			expectedErr: create.ErrInvalidName,
		},
		{
			name:           "returns error when uniqueness validation fails",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockEventPublisher *mocks.MockEventDispatcher, mockLogger *loggermocks.MockLogger, input CreateSpecialistDTO) chan struct{} {
				mockRepo.EXPECT().SaveWithValidation(gomock.Any(), gomock.Any()).Return(nil, create.ErrDuplicateEmail).Times(1)
				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
				return nil
			},
			expectError: true,
			expectedErr: create.ErrDuplicateEmail,
		},
		{
			name:           "returns error when repository save fails",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockEventPublisher *mocks.MockEventDispatcher, mockLogger *loggermocks.MockLogger, input CreateSpecialistDTO) chan struct{} {
				mockRepo.EXPECT().SaveWithValidation(gomock.Any(), gomock.Any()).Return(nil, errors.New("db connection lost")).Times(1)
				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
				return nil
			},
			expectError: true,
		},
		{
			name:           "succeeds and logs error when event publish fails",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockEventPublisher *mocks.MockEventDispatcher, mockLogger *loggermocks.MockLogger, input CreateSpecialistDTO) chan struct{} {
				done := make(chan struct{})

				mockRepo.EXPECT().SaveWithValidation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error) {
					return specialist, nil
				}).Times(1)

				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, evt interface{}) error {
					return errors.New("sqs unavailable")
				}).Times(1)

				mockLogger.EXPECT().Error(gomock.Any(), gomock.Eq(ErrEventPublishMessage), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, msg string, fields ...interface{}) {
					defer close(done)
				}).Times(1)

				return done
			},
			expectError: false,
			validateResult: func(t *testing.T, specialist *domain.Specialist) {
				assert.NotNil(t, specialist)
				assert.Equal(t, domain.StatusPending, specialist.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			input := createSpecialistFactory(tt.inputOverrides...)

			mockRepo := mocks.NewMockSpecialistCreateRepositoryInterface(ctrl)
			mockEventPublisher := mocks.NewMockEventDispatcher(ctrl)
			mockLogger := loggermocks.NewMockLogger(ctrl)

			done := tt.setupMocks(mockRepo, mockEventPublisher, mockLogger, *input)

			useCase := NewCreateSpecialistUseCase(
				mockRepo,
				mockEventPublisher,
				mockLogger,
			)

			result, err := useCase.Execute(context.Background(), *input)

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

			if done != nil {
				select {
				case <-done:
				case <-time.After(2 * time.Second):
					t.Fatal("timed out waiting for async event dispatch")
				}
			}
		})
	}
}
