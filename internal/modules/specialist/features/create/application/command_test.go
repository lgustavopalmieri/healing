package application

import (
	"context"
	"errors"
	"testing"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/create"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application/mocks"
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

func TestCreateSpecialistCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		inputOverrides []func(*CreateSpecialistDTO)
		setupMocks     func(*mocks.MockSpecialistCreateRepositoryInterface, *mocks.MockEventDispatcher, *mocks.MockTracer, *mocks.MockLogger, *mocks.MockSpan, CreateSpecialistDTO)
		expectError    bool
		expectedErr    error
		validateResult func(*testing.T, *domain.Specialist)
	}{
		{
			name:           "creates specialist with status pending and publishes event",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockEventPublisher *mocks.MockEventDispatcher, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan, input CreateSpecialistDTO) {
				ctx := context.Background()

				mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(nil).Times(1)

				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error) {
					return specialist, nil
				}).Times(1)

				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Return(nil).Times(1)
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
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockEventPublisher *mocks.MockEventDispatcher, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan, input CreateSpecialistDTO) {
				ctx := context.Background()

				mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)
				mockSpan.EXPECT().RecordError(create.ErrInvalidName).Times(1)
				mockLogger.EXPECT().Error(gomock.Any(), create.ErrInvalidName.Error(), gomock.Any()).Times(1)

				mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: create.ErrInvalidName,
		},
		{
			name:           "returns error when uniqueness validation fails",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockEventPublisher *mocks.MockEventDispatcher, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan, input CreateSpecialistDTO) {
				ctx := context.Background()

				mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(create.ErrDuplicateEmail).Times(1)
				mockSpan.EXPECT().RecordError(create.ErrDuplicateEmail).Times(1)
				mockLogger.EXPECT().Error(gomock.Any(), ErrUniquenessValidationMessage, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: create.ErrDuplicateEmail,
		},
		{
			name:           "returns error when repository save fails",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockEventPublisher *mocks.MockEventDispatcher, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan, input CreateSpecialistDTO) {
				ctx := context.Background()

				mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(nil).Times(1)

				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil, errors.New("db connection lost")).Times(1)
				mockSpan.EXPECT().RecordError(gomock.Any()).Times(1)
				mockLogger.EXPECT().Error(gomock.Any(), ErrSaveSpecialistMessage, gomock.Any(), gomock.Any()).Times(1)

				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrSaveSpecialist,
		},
		{
			name:           "still succeeds when event publish fails",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockEventPublisher *mocks.MockEventDispatcher, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan, input CreateSpecialistDTO) {
				ctx := context.Background()

				mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(nil).Times(1)

				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error) {
					return specialist, nil
				}).Times(1)

				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Return(errors.New("kafka unavailable")).Times(1)
				mockLogger.EXPECT().Warn(gomock.Any(), ErrEventPublishMessage, gomock.Any(), gomock.Any()).Times(1)
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
			mockTracer := mocks.NewMockTracer(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockSpan := mocks.NewMockSpan(ctrl)

			tt.setupMocks(mockRepo, mockEventPublisher, mockTracer, mockLogger, mockSpan, *input)

			command := NewCreateSpecialistCommand(
				mockRepo,
				mockEventPublisher,
				mockTracer,
				mockLogger,
			)

			result, err := command.Execute(context.Background(), *input)

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
