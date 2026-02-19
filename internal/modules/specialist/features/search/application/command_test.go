package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func searchInputFactory(overrides ...func(*searchinput.ListSearchInput)) *searchinput.ListSearchInput {
	searchTerm := "cardiologia"
	pagination, _ := cursor.NewCursorPaginationInput(nil, 20, cursor.DirectionNext)
	input := &searchinput.ListSearchInput{
		SearchTerm: &searchTerm,
		Filters:    []searchinput.Filter{},
		Sort:       []searchinput.Sort{},
		Pagination: pagination,
	}
	for _, override := range overrides {
		override(input)
	}
	return input
}

func specialistFactory(overrides ...func(*domain.Specialist)) *domain.Specialist {
	now := time.Now()
	specialist := &domain.Specialist{
		ID:            "123e4567-e89b-12d3-a456-426614174000",
		Name:          "Dr. João Silva",
		Email:         "joao@example.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiologia",
		LicenseNumber: "CRM123456",
		Description:   "Especialista em cardiologia",
		Keywords:      []string{"coração", "arritmia"},
		AgreedToShare: true,
		Rating:        4.5,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, override := range overrides {
		override(specialist)
	}
	return specialist
}

func TestSearchSpecialistsCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		inputOverrides []func(*searchinput.ListSearchInput)
		setupMocks     func(*mocks.MockSpecialistSearchRepositoryInterface, *mocks.MockLogger, *searchinput.ListSearchInput)
		expectError    bool
		expectedErr    error
		validateResult func(*testing.T, *searchoutput.ListSearchOutput)
	}{
		{
			name:           "successfully executes search and returns results with specialists",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistSearchRepositoryInterface, mockLogger *mocks.MockLogger, input *searchinput.ListSearchInput) {
				specialists := []*domain.Specialist{
					specialistFactory(),
					specialistFactory(func(s *domain.Specialist) {
						s.ID = "223e4567-e89b-12d3-a456-426614174001"
						s.Name = "Dr. Maria Santos"
						s.Email = "maria@example.com"
					}),
				}

				nextCursor := "encoded_cursor_next"
				cursorOutput := cursor.NewCursorPaginationOutput(
					&nextCursor,
					nil,
					true,
					false,
					2,
				)

				output := searchoutput.NewListSearchOutput(specialists, cursorOutput)

				mockRepo.EXPECT().
					Search(gomock.Any(), input).
					Return(output, nil).
					Times(1)

				mockLogger.EXPECT().
					Error(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)

				mockLogger.EXPECT().
					Info(gomock.Any(), SearchNoResultsMessage).
					Times(0)
			},
			expectError: false,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				assert.NotNil(t, output)
				assert.Len(t, output.Specialists, 2)
				assert.Equal(t, "Dr. João Silva", output.Specialists[0].Name)
				assert.Equal(t, "Dr. Maria Santos", output.Specialists[1].Name)
				assert.True(t, output.CursorOutput.HasNextPage)
				assert.False(t, output.CursorOutput.HasPreviousPage)
				assert.Equal(t, 2, output.CursorOutput.TotalItemsInPage)
			},
		},
		{
			name:           "successfully executes search and returns empty results",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistSearchRepositoryInterface, mockLogger *mocks.MockLogger, input *searchinput.ListSearchInput) {
				specialists := []*domain.Specialist{}

				cursorOutput := cursor.NewCursorPaginationOutput(
					nil,
					nil,
					false,
					false,
					0,
				)

				output := searchoutput.NewListSearchOutput(specialists, cursorOutput)

				mockRepo.EXPECT().
					Search(gomock.Any(), input).
					Return(output, nil).
					Times(1)

				mockLogger.EXPECT().
					Info(gomock.Any(), SearchNoResultsMessage).
					Times(1)

				mockLogger.EXPECT().
					Error(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectError: false,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				assert.NotNil(t, output)
				assert.Empty(t, output.Specialists)
				assert.False(t, output.CursorOutput.HasNextPage)
				assert.False(t, output.CursorOutput.HasPreviousPage)
				assert.Equal(t, 0, output.CursorOutput.TotalItemsInPage)
				assert.True(t, output.CursorOutput.IsEmpty())
			},
		},
		{
			name:           "logs info message when no results are found",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistSearchRepositoryInterface, mockLogger *mocks.MockLogger, input *searchinput.ListSearchInput) {
				specialists := []*domain.Specialist{}

				cursorOutput := cursor.NewCursorPaginationOutput(
					nil,
					nil,
					false,
					false,
					0,
				)

				output := searchoutput.NewListSearchOutput(specialists, cursorOutput)

				mockRepo.EXPECT().
					Search(gomock.Any(), input).
					Return(output, nil).
					Times(1)

				mockLogger.EXPECT().
					Info(gomock.Any(), SearchNoResultsMessage).
					Times(1)
			},
			expectError: false,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				assert.NotNil(t, output)
				assert.True(t, output.CursorOutput.IsEmpty())
			},
		},
		{
			name:           "returns error when repository search fails",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistSearchRepositoryInterface, mockLogger *mocks.MockLogger, input *searchinput.ListSearchInput) {
				repoError := errors.New("elasticsearch connection failed")

				mockRepo.EXPECT().
					Search(gomock.Any(), input).
					Return(nil, repoError).
					Times(1)

				mockLogger.EXPECT().
					Error(gomock.Any(), ErrSearchExecutionMessage, gomock.Any()).
					Times(1)

				mockLogger.EXPECT().
					Info(gomock.Any(), SearchNoResultsMessage).
					Times(0)
			},
			expectError: true,
			expectedErr: ErrSearchExecution,
		},
		{
			name:           "logs error message when repository search fails",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistSearchRepositoryInterface, mockLogger *mocks.MockLogger, input *searchinput.ListSearchInput) {
				repoError := errors.New("index not found")

				mockRepo.EXPECT().
					Search(gomock.Any(), input).
					Return(nil, repoError).
					Times(1)

				mockLogger.EXPECT().
					Error(gomock.Any(), ErrSearchExecutionMessage, gomock.Any()).
					Times(1)
			},
			expectError: true,
			expectedErr: ErrSearchExecution,
		},
		{
			name:           "propagates context cancellation from repository",
			inputOverrides: nil,
			setupMocks: func(mockRepo *mocks.MockSpecialistSearchRepositoryInterface, mockLogger *mocks.MockLogger, input *searchinput.ListSearchInput) {
				ctxErr := context.Canceled

				mockRepo.EXPECT().
					Search(gomock.Any(), input).
					Return(nil, ctxErr).
					Times(1)

				mockLogger.EXPECT().
					Error(gomock.Any(), ErrSearchExecutionMessage, gomock.Any()).
					Times(1)

				mockLogger.EXPECT().
					Info(gomock.Any(), SearchNoResultsMessage).
					Times(0)
			},
			expectError: true,
			expectedErr: ErrSearchExecution,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			input := searchInputFactory(tt.inputOverrides...)

			mockRepo := mocks.NewMockSpecialistSearchRepositoryInterface(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			tt.setupMocks(mockRepo, mockLogger, input)

			command := NewSearchSpecialistsCommand(mockRepo, mockLogger)

			ctx := context.Background()
			result, err := command.Execute(ctx, input)

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
