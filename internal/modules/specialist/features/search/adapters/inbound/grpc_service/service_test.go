package grpcservice

import (
	"context"
	"testing"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/inbound/grpc_service/mocks"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/inbound/grpc_service/pb"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source=service.go -destination=mocks/command_mock.go -package=mocks

func searchRequestFactory(overrides ...func(*pb.SearchSpecialistsRequest)) *pb.SearchSpecialistsRequest {
	req := &pb.SearchSpecialistsRequest{
		SearchTerm: "cardiology",
		Filters:    []*pb.SearchFilter{},
		Sort: []*pb.SortCriteria{
			{Field: "rating", Order: "desc"},
		},
		PageSize:  10,
		Cursor:    "",
		Direction: "next",
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
		Rating:        4.8,
		Status:        domain.StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	for _, override := range overrides {
		override(specialist)
	}

	return specialist
}

func searchOutputFactory(specialists []*domain.Specialist, nextCursor *string, prevCursor *string, hasNext bool, hasPrev bool) *searchoutput.ListSearchOutput {
	return searchoutput.NewListSearchOutput(
		specialists,
		cursor.NewCursorPaginationOutput(nextCursor, prevCursor, hasNext, hasPrev, len(specialists)),
	)
}

func TestSpecialistSearchGRPCService_SearchSpecialists(t *testing.T) {
	tests := []struct {
		name             string
		input            *pb.SearchSpecialistsRequest
		setupContext     func() context.Context
		setupMocks       func(*mocks.MockSpecialistSearchCommandInterface)
		expectError      bool
		expectedErr      error
		validateResponse func(*testing.T, *pb.SearchSpecialistsResponse)
	}{
		{
			name:  "success - searches specialists with search term and returns paginated results",
			input: searchRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				specialists := []*domain.Specialist{specialistFactory()}
				output := searchOutputFactory(specialists, nil, nil, false, false)
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(output, nil).
					Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResponse: func(t *testing.T, response *pb.SearchSpecialistsResponse) {
				assert.NotNil(t, response)
				assert.Len(t, response.Specialists, 1)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", response.Specialists[0].Id)
				assert.Equal(t, "Dr. João Silva", response.Specialists[0].Name)
				assert.Equal(t, "joao@exemplo.com", response.Specialists[0].Email)
				assert.Equal(t, "Cardiology", response.Specialists[0].Specialty)
				assert.Equal(t, "CRM-123456", response.Specialists[0].LicenseNumber)
				assert.Equal(t, []string{"heart", "cardiology"}, response.Specialists[0].Keywords)
				assert.True(t, response.Specialists[0].AgreedToShare)
				assert.Equal(t, 4.8, response.Specialists[0].Rating)
				assert.Equal(t, "active", response.Specialists[0].Status)
				assert.NotNil(t, response.Pagination)
				assert.False(t, response.Pagination.HasNextPage)
				assert.False(t, response.Pagination.HasPreviousPage)
				assert.Equal(t, int32(1), response.Pagination.TotalItemsInPage)
			},
		},
		{
			name: "success - searches specialists with filters only and returns results",
			input: searchRequestFactory(func(req *pb.SearchSpecialistsRequest) {
				req.SearchTerm = ""
				req.Filters = []*pb.SearchFilter{
					{Field: "specialty", Value: "Cardiology"},
				}
			}),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				specialists := []*domain.Specialist{
					specialistFactory(),
					specialistFactory(func(s *domain.Specialist) {
						s.ID = "660e8400-e29b-41d4-a716-446655440001"
						s.Name = "Dr. Maria Santos"
						s.Email = "maria@exemplo.com"
					}),
				}
				output := searchOutputFactory(specialists, nil, nil, false, false)
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(output, nil).
					Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResponse: func(t *testing.T, response *pb.SearchSpecialistsResponse) {
				assert.NotNil(t, response)
				assert.Len(t, response.Specialists, 2)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", response.Specialists[0].Id)
				assert.Equal(t, "660e8400-e29b-41d4-a716-446655440001", response.Specialists[1].Id)
			},
		},
		{
			name:  "success - returns empty result when no specialists match",
			input: searchRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				output := searchOutputFactory([]*domain.Specialist{}, nil, nil, false, false)
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(output, nil).
					Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResponse: func(t *testing.T, response *pb.SearchSpecialistsResponse) {
				assert.NotNil(t, response)
				assert.Empty(t, response.Specialists)
				assert.NotNil(t, response.Pagination)
				assert.Equal(t, int32(0), response.Pagination.TotalItemsInPage)
				assert.False(t, response.Pagination.HasNextPage)
			},
		},
		{
			name:  "success - returns response with next page cursor when has more results",
			input: searchRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				specialists := []*domain.Specialist{specialistFactory()}
				nextCur := "eyJzb3J0IjpbNC44LCIyMDI0LTAxLTAxVDAwOjAwOjAwWiJdfQ=="
				output := searchOutputFactory(specialists, &nextCur, nil, true, false)
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(output, nil).
					Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResponse: func(t *testing.T, response *pb.SearchSpecialistsResponse) {
				assert.NotNil(t, response)
				assert.Len(t, response.Specialists, 1)
				assert.True(t, response.Pagination.HasNextPage)
				assert.False(t, response.Pagination.HasPreviousPage)
				assert.NotEmpty(t, response.Pagination.NextCursor)
				assert.Empty(t, response.Pagination.PreviousCursor)
			},
		},
		{
			name: "success - searches with cursor for second page navigation",
			input: searchRequestFactory(func(req *pb.SearchSpecialistsRequest) {
				req.Cursor = "eyJzb3J0IjpbNC44LCIyMDI0LTAxLTAxVDAwOjAwOjAwWiJdfQ=="
				req.Direction = "next"
			}),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				specialists := []*domain.Specialist{specialistFactory(func(s *domain.Specialist) {
					s.ID = "770e8400-e29b-41d4-a716-446655440002"
					s.Name = "Dr. Pedro Costa"
				})}
				prevCur := "eyJzb3J0IjpbNC41LCIyMDI0LTAxLTAyVDAwOjAwOjAwWiJdfQ=="
				output := searchOutputFactory(specialists, nil, &prevCur, false, true)
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(output, nil).
					Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResponse: func(t *testing.T, response *pb.SearchSpecialistsResponse) {
				assert.NotNil(t, response)
				assert.Len(t, response.Specialists, 1)
				assert.Equal(t, "770e8400-e29b-41d4-a716-446655440002", response.Specialists[0].Id)
				assert.False(t, response.Pagination.HasNextPage)
				assert.True(t, response.Pagination.HasPreviousPage)
				assert.NotEmpty(t, response.Pagination.PreviousCursor)
			},
		},
		{
			name: "failure - propagates ErrInvalidSearchInput when command rejects empty criteria",
			input: searchRequestFactory(func(req *pb.SearchSpecialistsRequest) {
				req.SearchTerm = ""
				req.Filters = []*pb.SearchFilter{}
			}),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, application.ErrInvalidSearchInput).
					Times(1)
			},
			expectError: true,
			expectedErr: application.ErrInvalidSearchInput,
			validateResponse: func(t *testing.T, response *pb.SearchSpecialistsResponse) {
				assert.Nil(t, response)
			},
		},
		{
			name:  "failure - propagates ErrInvalidSearchInput from command",
			input: searchRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, application.ErrInvalidSearchInput).
					Times(1)
			},
			expectError: true,
			expectedErr: application.ErrInvalidSearchInput,
			validateResponse: func(t *testing.T, response *pb.SearchSpecialistsResponse) {
				assert.Nil(t, response)
			},
		},
		{
			name:  "failure - propagates ErrSearchExecution from command",
			input: searchRequestFactory(),
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, application.ErrSearchExecution).
					Times(1)
			},
			expectError: true,
			expectedErr: application.ErrSearchExecution,
			validateResponse: func(t *testing.T, response *pb.SearchSpecialistsResponse) {
				assert.Nil(t, response)
			},
		},
		{
			name:  "failure - handles context cancellation gracefully",
			input: searchRequestFactory(),
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, context.Canceled).
					Times(1)
			},
			expectError: true,
			expectedErr: context.Canceled,
			validateResponse: func(t *testing.T, response *pb.SearchSpecialistsResponse) {
				assert.Nil(t, response)
			},
		},
		{
			name:  "failure - returns ErrNilRequest when request is nil",
			input: nil,
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func(mockCommand *mocks.MockSpecialistSearchCommandInterface) {
				mockCommand.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectError: true,
			expectedErr: ErrNilRequest,
			validateResponse: func(t *testing.T, response *pb.SearchSpecialistsResponse) {
				assert.Nil(t, response)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCommand := mocks.NewMockSpecialistSearchCommandInterface(ctrl)
			tt.setupMocks(mockCommand)

			service := NewSpecialistSearchGRPCService(mockCommand)
			ctx := tt.setupContext()

			response, err := service.SearchSpecialists(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}

			tt.validateResponse(t, response)
		})
	}
}
