package application

import (
	"context"
	"errors"
	"testing"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
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
		setupMocks     func(*mocks.MockSpecialistCreateRepositoryInterface, *mocks.MockSpecialistCreateExternalGatewayInterface, *mocks.MockEventDispatcher, *mocks.MockTracer, *mocks.MockLogger, *mocks.MockSpan, *mocks.MockSpan, CreateSpecialistDTO)
		expectError    bool
		expectedErr    error
		validateResult func(*testing.T, *domain.Specialist)
	}{
		{
			name:           "1-validate license number - success case",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockGateway *mocks.MockSpecialistCreateExternalGatewayInterface, mockEventPublisher *mocks.MockEventDispatcher, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan, mockApiSpan *mocks.MockSpan, input CreateSpecialistDTO) {
				ctx := context.Background()

				mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(nil).Times(1)

				mockTracer.EXPECT().Start(gomock.Any(), "ValidateLicenseExternal").Return(ctx, mockApiSpan).Times(1)
				mockApiSpan.EXPECT().End().Times(1)
				mockGateway.EXPECT().ValidateLicenseNumber(gomock.Any(), input.LicenseNumber).Return(true, nil).Times(1)

				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error) {
					return specialist, nil
				}).Times(1)

				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				mockLogger.EXPECT().Info(gomock.Any(), SpecialistCreatedSuccessMessage, gomock.Any(), gomock.Any()).Times(1)
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
				assert.NotEmpty(t, specialist.ID)
				assert.False(t, specialist.CreatedAt.IsZero())
				assert.False(t, specialist.UpdatedAt.IsZero())
			},
		},
		{
			name:           "2-invalid license number - fail case",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockGateway *mocks.MockSpecialistCreateExternalGatewayInterface, mockEventPublisher *mocks.MockEventDispatcher, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan, mockApiSpan *mocks.MockSpan, input CreateSpecialistDTO) {
				ctx := context.Background()

				mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(nil).Times(1)

				mockTracer.EXPECT().Start(gomock.Any(), "ValidateLicenseExternal").Return(ctx, mockApiSpan).Times(1)
				mockApiSpan.EXPECT().End().Times(1)
				mockGateway.EXPECT().ValidateLicenseNumber(gomock.Any(), input.LicenseNumber).Return(false, nil).Times(1)

				mockLogger.EXPECT().Warn(gomock.Any(), InvalidLicenseNumberMessage, gomock.Any()).Times(1)

				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
				mockLogger.EXPECT().Info(gomock.Any(), SpecialistCreatedSuccessMessage, gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrInvalidLicense,
		},
		{
			name:           "3-external gateway error - fail case",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistCreateRepositoryInterface, mockGateway *mocks.MockSpecialistCreateExternalGatewayInterface, mockEventPublisher *mocks.MockEventDispatcher, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan, mockApiSpan *mocks.MockSpan, input CreateSpecialistDTO) {
				ctx := context.Background()
				gatewayError := errors.New("server is down")

				mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(nil).Times(1)

				mockTracer.EXPECT().Start(gomock.Any(), "ValidateLicenseExternal").Return(ctx, mockApiSpan).Times(1)
				mockApiSpan.EXPECT().End().Times(1)
				mockApiSpan.EXPECT().RecordError(gatewayError).Times(1)
				mockGateway.EXPECT().ValidateLicenseNumber(gomock.Any(), input.LicenseNumber).Return(false, gatewayError).Times(1)

				mockLogger.EXPECT().Error(gomock.Any(), ErrLicenseValidationMessage, gomock.Any(), gomock.Any()).Times(1)

				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
				mockLogger.EXPECT().Info(gomock.Any(), SpecialistCreatedSuccessMessage, gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrLicenseValidation,
		},
		{
			name:           "4-external validation timeout",
			inputOverrides: nil,
			setupMocks: func(
				mockRepo *mocks.MockSpecialistCreateRepositoryInterface,
				mockGateway *mocks.MockSpecialistCreateExternalGatewayInterface,
				mockEventPublisher *mocks.MockEventDispatcher,
				mockTracer *mocks.MockTracer,
				mockLogger *mocks.MockLogger,
				mockSpan *mocks.MockSpan,
				mockApiSpan *mocks.MockSpan,
				input CreateSpecialistDTO,
			) {
				ctx := context.Background()

				mockTracer.EXPECT().
					Start(gomock.Any(), CreateSpecialistSpanName).
					Return(ctx, mockSpan).
					Times(1)

				mockSpan.EXPECT().End().Times(1)

				mockRepo.EXPECT().
					ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).
					Return(nil).
					Times(1)

				mockTracer.EXPECT().
					Start(gomock.Any(), "ValidateLicenseExternal").
					Return(ctx, mockApiSpan).
					AnyTimes()

				mockApiSpan.EXPECT().End().
					AnyTimes()

				mockGateway.EXPECT().
					ValidateLicenseNumber(gomock.Any(), input.LicenseNumber).
					DoAndReturn(func(ctx context.Context, _ string) (bool, error) {
						<-ctx.Done()
						return false, ctx.Err()
					}).
					Times(1)

				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
				mockLogger.EXPECT().Info(gomock.Any(), SpecialistCreatedSuccessMessage, gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrExternalValidationTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			input := createSpecialistFactory(tt.inputOverrides...)

			mockRepo := mocks.NewMockSpecialistCreateRepositoryInterface(ctrl)
			mockGateway := mocks.NewMockSpecialistCreateExternalGatewayInterface(ctrl)
			mockEventPublisher := mocks.NewMockEventDispatcher(ctrl)
			mockTracer := mocks.NewMockTracer(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockSpan := mocks.NewMockSpan(ctrl)
			mockApiSpan := mocks.NewMockSpan(ctrl)

			tt.setupMocks(mockRepo, mockGateway, mockEventPublisher, mockTracer, mockLogger, mockSpan, mockApiSpan, *input)

			command := NewCreateSpecialistCommand(
				mockRepo,
				mockGateway,
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
