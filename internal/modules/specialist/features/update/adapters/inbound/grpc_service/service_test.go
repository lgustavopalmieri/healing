package grpcservice

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/adapters/inbound/grpc_service/mocks"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/adapters/inbound/grpc_service/pb"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -source=service.go -destination=mocks/usecase_mock.go -package=mocks

func updateRequestFactory(overrides ...func(*pb.UpdateSpecialistRequest)) *pb.UpdateSpecialistRequest {
	req := &pb.UpdateSpecialistRequest{
		Id:            "550e8400-e29b-41d4-a716-446655440000",
		Name:          wrapperspb.String("Dr. Maria Santos"),
		Email:         wrapperspb.String("maria@example.com"),
		Phone:         wrapperspb.String("+5511888888888"),
		Specialty:     wrapperspb.String("Neurologia"),
		LicenseNumber: wrapperspb.String("CRM654321"),
		Description:   wrapperspb.String("Neurologista experiente"),
		Keywords:      []string{"neurologia", "cérebro"},
		AgreedToShare: wrapperspb.Bool(true),
		Status:        wrapperspb.String("active"),
	}
	for _, o := range overrides {
		o(req)
	}
	return req
}

func specialistFactory(overrides ...func(*domain.Specialist)) *domain.Specialist {
	now := time.Now().UTC()
	s := &domain.Specialist{
		ID:            "550e8400-e29b-41d4-a716-446655440000",
		Name:          "Dr. Maria Santos",
		Email:         "maria@example.com",
		Phone:         "+5511888888888",
		Specialty:     "Neurologia",
		LicenseNumber: "CRM654321",
		Description:   "Neurologista experiente",
		Keywords:      []string{"neurologia", "cérebro"},
		AgreedToShare: true,
		Rating:        4.8,
		Status:        domain.StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, o := range overrides {
		o(s)
	}
	return s
}

func TestSpecialistUpdateGRPCService_UpdateSpecialist(t *testing.T) {
	tests := []struct {
		name             string
		input            *pb.UpdateSpecialistRequest
		setupContext     func() context.Context
		setupMocks     func(*mocks.MockSpecialistUpdateUseCaseInterface)
		expectError      bool
		expectedErr      error
		validateResponse func(*testing.T, *pb.UpdateSpecialistResponse)
	}{
		{
			name:  "success - updates specialist successfully with all fields",
			input: updateRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockUseCase *mocks.MockSpecialistUpdateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dto application.UpdateSpecialistDTO) (*domain.Specialist, error) {
						assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", dto.ID)
						assert.Equal(t, "Dr. Maria Santos", *dto.Name)
						assert.Equal(t, "maria@example.com", *dto.Email)
						assert.Equal(t, "+5511888888888", *dto.Phone)
						assert.Equal(t, "Neurologia", *dto.Specialty)
						assert.Equal(t, "CRM654321", *dto.LicenseNumber)
						assert.Equal(t, "Neurologista experiente", *dto.Description)
						assert.Equal(t, []string{"neurologia", "cérebro"}, dto.Keywords)
						assert.True(t, *dto.AgreedToShare)
						assert.Equal(t, domain.StatusActive, *dto.Status)
						return specialistFactory(), nil
					}).
					Times(1)
			},
			expectError: false,
			validateResponse: func(t *testing.T, resp *pb.UpdateSpecialistResponse) {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Specialist)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", resp.Specialist.Id)
				assert.Equal(t, "Dr. Maria Santos", resp.Specialist.Name)
				assert.Equal(t, "maria@example.com", resp.Specialist.Email)
				assert.Equal(t, "+5511888888888", resp.Specialist.Phone)
				assert.Equal(t, "Neurologia", resp.Specialist.Specialty)
				assert.Equal(t, "CRM654321", resp.Specialist.LicenseNumber)
				assert.Equal(t, "Neurologista experiente", resp.Specialist.Description)
				assert.Equal(t, []string{"neurologia", "cérebro"}, resp.Specialist.Keywords)
				assert.True(t, resp.Specialist.AgreedToShare)
				assert.Equal(t, 4.8, resp.Specialist.Rating)
				assert.Equal(t, "active", resp.Specialist.Status)
				assert.NotNil(t, resp.Specialist.CreatedAt)
				assert.NotNil(t, resp.Specialist.UpdatedAt)
			},
		},
		{
			name: "success - updates specialist with partial fields (only name)",
			input: updateRequestFactory(func(req *pb.UpdateSpecialistRequest) {
				req.Email = nil
				req.Phone = nil
				req.Specialty = nil
				req.LicenseNumber = nil
				req.Description = nil
				req.Keywords = nil
				req.AgreedToShare = nil
				req.Status = nil
			}),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockUseCase *mocks.MockSpecialistUpdateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dto application.UpdateSpecialistDTO) (*domain.Specialist, error) {
						assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", dto.ID)
						assert.NotNil(t, dto.Name)
						assert.Equal(t, "Dr. Maria Santos", *dto.Name)
						assert.Nil(t, dto.Email)
						assert.Nil(t, dto.Phone)
						assert.Nil(t, dto.Specialty)
						assert.Nil(t, dto.LicenseNumber)
						assert.Nil(t, dto.Description)
						assert.Nil(t, dto.Keywords)
						assert.Nil(t, dto.AgreedToShare)
						assert.Nil(t, dto.Status)
						return specialistFactory(), nil
					}).
					Times(1)
			},
			expectError: false,
			validateResponse: func(t *testing.T, resp *pb.UpdateSpecialistResponse) {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Specialist)
				assert.Equal(t, "Dr. Maria Santos", resp.Specialist.Name)
			},
		},
		{
			name:  "failure - propagates ErrSpecialistNotFound from command",
			input: updateRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockUseCase *mocks.MockSpecialistUpdateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, application.ErrSpecialistNotFound).
					Times(1)
			},
			expectError: true,
			expectedErr: application.ErrSpecialistNotFound,
			validateResponse: func(t *testing.T, resp *pb.UpdateSpecialistResponse) {
				assert.Nil(t, resp)
			},
		},
		{
			name: "failure - propagates domain ErrInvalidName from command",
			input: updateRequestFactory(func(req *pb.UpdateSpecialistRequest) {
				req.Name = wrapperspb.String("")
			}),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockUseCase *mocks.MockSpecialistUpdateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrInvalidName).
					Times(1)
			},
			expectError: true,
			expectedErr: domain.ErrInvalidName,
			validateResponse: func(t *testing.T, resp *pb.UpdateSpecialistResponse) {
				assert.Nil(t, resp)
			},
		},
		{
			name:  "failure - propagates ErrUpdateSpecialist from command",
			input: updateRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockUseCase *mocks.MockSpecialistUpdateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, application.ErrUpdateSpecialist).
					Times(1)
			},
			expectError: true,
			expectedErr: application.ErrUpdateSpecialist,
			validateResponse: func(t *testing.T, resp *pb.UpdateSpecialistResponse) {
				assert.Nil(t, resp)
			},
		},
		{
			name:  "failure - handles context cancellation gracefully",
			input: updateRequestFactory(),
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupMocks: func(mockUseCase *mocks.MockSpecialistUpdateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, context.Canceled).
					Times(1)
			},
			expectError: true,
			expectedErr: context.Canceled,
			validateResponse: func(t *testing.T, resp *pb.UpdateSpecialistResponse) {
				assert.Nil(t, resp)
			},
		},
		{
			name:  "failure - handles nil request by converting to empty DTO",
			input: nil,
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockUseCase *mocks.MockSpecialistUpdateUseCaseInterface) {
				mockUseCase.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dto application.UpdateSpecialistDTO) (*domain.Specialist, error) {
						assert.Empty(t, dto.ID)
						assert.Nil(t, dto.Name)
						assert.Nil(t, dto.Email)
						assert.Nil(t, dto.Phone)
						assert.Nil(t, dto.Specialty)
						assert.Nil(t, dto.LicenseNumber)
						assert.Nil(t, dto.Description)
						assert.Nil(t, dto.Keywords)
						assert.Nil(t, dto.AgreedToShare)
						assert.Nil(t, dto.Status)
						return nil, application.ErrSpecialistNotFound
					}).
					Times(1)
			},
			expectError: true,
			expectedErr: application.ErrSpecialistNotFound,
			validateResponse: func(t *testing.T, resp *pb.UpdateSpecialistResponse) {
				assert.Nil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mocks.NewMockSpecialistUpdateUseCaseInterface(ctrl)
			tt.setupMocks(mockUseCase)

			service := NewSpecialistUpdateGRPCService(mockUseCase)
			ctx := tt.setupContext()

			resp, err := service.UpdateSpecialist(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResponse(t, resp)
		})
	}
}
