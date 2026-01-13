package grpchandler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/grpc_handler/mocks"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/grpc_handler/pb"
)

//go:generate mockgen -source=handler.go -destination=mocks/command_mock.go -package=mocks

// go test ./internal/modules/specialist/features/create/infra/grpc_handler/ -v
// go test ./internal/modules/specialist/features/create/infra/grpc_handler/ -cover
func createSpecialistRequestFactory(overrides ...func(*pb.CreateSpecialistRequest)) *pb.CreateSpecialistRequest {
	req := &pb.CreateSpecialistRequest{
		Name:          "Dr. João Silva",
		Email:         "joao@exemplo.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiology",
		LicenseNumber: "CRM-123456",
		Description:   "Especialista em cardiologia clínica",
		Keywords:      []string{"heart", "cardiology"},
		AgreedToShare: true,
	}

	for _, override := range overrides {
		override(req)
	}

	return req
}

func specialistFactory(overrides ...func(*domain.Specialist)) *domain.Specialist {
	now := time.Now().UTC()
	specialist := &domain.Specialist{
		ID:            "550e8400-e29b-41d4-a716-446655440000",
		Name:          "Dr. João Silva",
		Email:         "joao@exemplo.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiology",
		LicenseNumber: "CRM-123456",
		Description:   "Especialista em cardiologia clínica",
		Keywords:      []string{"heart", "cardiology"},
		AgreedToShare: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	for _, override := range overrides {
		override(specialist)
	}

	return specialist
}

func TestSpecialistCreateGRPCHandler_Handle(t *testing.T) {
	tests := []struct {
		name             string
		input            *pb.CreateSpecialistRequest
		setupContext     func() context.Context
		setupMocks       func(*mocks.MockSpecialistCreateCommandInterface)
		expectError      bool
		expectedErr      error
		validateResponse func(*testing.T, *pb.CreateSpecialistResponse)
	}{
		{
			name:  "success - creates specialist successfully with all valid data",
			input: createSpecialistRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistCreateCommandInterface) {
				expectedDTO := application.CreateSpecialistDTO{
					Name:          "Dr. João Silva",
					Email:         "joao@exemplo.com",
					Phone:         "+5511999999999",
					Specialty:     "Cardiology",
					LicenseNumber: "CRM-123456",
					Description:   "Especialista em cardiologia clínica",
					Keywords:      []string{"heart", "cardiology"},
					AgreedToShare: true,
				}
				mockCommand.EXPECT().
					Execute(gomock.Any(), expectedDTO).
					Return(specialistFactory(), nil).
					Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResponse: func(t *testing.T, response *pb.CreateSpecialistResponse) {
				assert.NotNil(t, response)
				assert.NotNil(t, response.Specialist)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", response.Specialist.Id)
				assert.Equal(t, "Dr. João Silva", response.Specialist.Name)
				assert.Equal(t, "joao@exemplo.com", response.Specialist.Email)
				assert.Equal(t, "+5511999999999", response.Specialist.Phone)
				assert.Equal(t, "Cardiology", response.Specialist.Specialty)
				assert.Equal(t, "CRM-123456", response.Specialist.LicenseNumber)
				assert.Equal(t, "Especialista em cardiologia clínica", response.Specialist.Description)
				assert.Equal(t, []string{"heart", "cardiology"}, response.Specialist.Keywords)
				assert.True(t, response.Specialist.AgreedToShare)
				assert.NotNil(t, response.Specialist.CreatedAt)
				assert.NotNil(t, response.Specialist.UpdatedAt)
			},
		},
		{
			name: "failure - propagates ErrInvalidName from command to handler",
			input: createSpecialistRequestFactory(func(req *pb.CreateSpecialistRequest) {
				req.Name = ""
			}),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistCreateCommandInterface) {
				expectedDTO := application.CreateSpecialistDTO{
					Name:          "",
					Email:         "joao@exemplo.com",
					Phone:         "+5511999999999",
					Specialty:     "Cardiology",
					LicenseNumber: "CRM-123456",
					Description:   "Especialista em cardiologia clínica",
					Keywords:      []string{"heart", "cardiology"},
					AgreedToShare: true,
				}
				mockCommand.EXPECT().
					Execute(gomock.Any(), expectedDTO).
					Return(nil, domain.ErrInvalidName).
					Times(1)
			},
			expectError: true,
			expectedErr: domain.ErrInvalidName,
			validateResponse: func(t *testing.T, response *pb.CreateSpecialistResponse) {
				assert.Nil(t, response)
			},
		},
		{
			name:  "failure - propagates ErrDuplicateEmail from command to handler",
			input: createSpecialistRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistCreateCommandInterface) {
				expectedDTO := application.CreateSpecialistDTO{
					Name:          "Dr. João Silva",
					Email:         "joao@exemplo.com",
					Phone:         "+5511999999999",
					Specialty:     "Cardiology",
					LicenseNumber: "CRM-123456",
					Description:   "Especialista em cardiologia clínica",
					Keywords:      []string{"heart", "cardiology"},
					AgreedToShare: true,
				}
				mockCommand.EXPECT().
					Execute(gomock.Any(), expectedDTO).
					Return(nil, domain.ErrDuplicateEmail).
					Times(1)
			},
			expectError: true,
			expectedErr: domain.ErrDuplicateEmail,
			validateResponse: func(t *testing.T, response *pb.CreateSpecialistResponse) {
				assert.Nil(t, response)
			},
		},
		{
			name:  "failure - propagates ErrExternalValidationTimeout from command to handler",
			input: createSpecialistRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistCreateCommandInterface) {
				expectedDTO := application.CreateSpecialistDTO{
					Name:          "Dr. João Silva",
					Email:         "joao@exemplo.com",
					Phone:         "+5511999999999",
					Specialty:     "Cardiology",
					LicenseNumber: "CRM-123456",
					Description:   "Especialista em cardiologia clínica",
					Keywords:      []string{"heart", "cardiology"},
					AgreedToShare: true,
				}
				mockCommand.EXPECT().
					Execute(gomock.Any(), expectedDTO).
					Return(nil, application.ErrExternalValidationTimeout).
					Times(1)
			},
			expectError: true,
			expectedErr: application.ErrExternalValidationTimeout,
			validateResponse: func(t *testing.T, response *pb.CreateSpecialistResponse) {
				assert.Nil(t, response)
			},
		},
		{
			name:  "failure - handles context cancellation gracefully",
			input: createSpecialistRequestFactory(),
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistCreateCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, context.Canceled).
					Times(1)
			},
			expectError: true,
			expectedErr: context.Canceled,
			validateResponse: func(t *testing.T, response *pb.CreateSpecialistResponse) {
				assert.Nil(t, response)
			},
		},
		{
			name:  "failure - handles nil request by converting to empty DTO and letting command validate",
			input: nil,
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistCreateCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dto application.CreateSpecialistDTO) (*domain.Specialist, error) {
						assert.Empty(t, dto.Name)
						assert.Empty(t, dto.Email)
						assert.Empty(t, dto.Phone)
						assert.Empty(t, dto.Specialty)
						assert.Empty(t, dto.LicenseNumber)
						assert.Empty(t, dto.Description)
						assert.Empty(t, dto.Keywords)
						assert.False(t, dto.AgreedToShare)
						return nil, domain.ErrInvalidName
					}).
					Times(1)
			},
			expectError: true,
			expectedErr: domain.ErrInvalidName,
			validateResponse: func(t *testing.T, response *pb.CreateSpecialistResponse) {
				assert.Nil(t, response)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCommand := mocks.NewMockSpecialistCreateCommandInterface(ctrl)
			tt.setupMocks(mockCommand)

			handler := NewSpecialistCreateGRPCHandler(mockCommand)
			ctx := tt.setupContext()

			response, err := handler.Handle(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResponse(t, response)
		})
	}
}
