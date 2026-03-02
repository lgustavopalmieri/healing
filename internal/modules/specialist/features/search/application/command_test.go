package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	loggerMocks "github.com/lgustavopalmieri/healing-specialist/internal/commom/observability/mocks"
	cursor "github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application/mocks"
)

func searchDTOFactory(overrides ...func(*SearchSpecialistsDTO)) *SearchSpecialistsDTO {
	searchTerm := "cardiologia"
	pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)

	dto := &SearchSpecialistsDTO{
		SearchTerm: &searchTerm,
		Filters:    nil,
		Sort:       nil,
		Pagination: pagination,
	}

	for _, override := range overrides {
		override(dto)
	}

	return dto
}

func searchResultFactory(overrides ...func(*searchoutput.SearchResult)) *searchoutput.SearchResult {
	now := time.Now()
	result := &searchoutput.SearchResult{
		Specialists: []*domain.Specialist{
			{
				ID:            uuid.New().String(),
				Name:          "Dr. João Silva",
				Email:         "joao@example.com",
				Phone:         "+5511999999999",
				Specialty:     "Cardiologia",
				LicenseNumber: "CRM-SP-123456",
				Description:   "Cardiologista experiente",
				Keywords:      []string{"cardiologia", "coração"},
				AgreedToShare: true,
				Rating:        4.8,
				Status:        domain.StatusActive,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		},
		HasNextPage:     false,
		FirstSortValues: []any{4.8, "2025-01-01T00:00:00Z", "abc-123"},
		LastSortValues:  []any{4.8, "2025-01-01T00:00:00Z", "abc-123"},
	}

	for _, override := range overrides {
		override(result)
	}

	return result
}

func emptySearchResult() *searchoutput.SearchResult {
	return &searchoutput.SearchResult{
		Specialists:     []*domain.Specialist{},
		HasNextPage:     false,
		FirstSortValues: nil,
		LastSortValues:  nil,
	}
}

func TestSearchSpecialistsCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		input          *SearchSpecialistsDTO
		setupMocks     func(*mocks.MockSpecialistSearchRepositoryInterface, *loggerMocks.MockLogger)
		expectError    bool
		expectedErr    error
		validateResult func(*testing.T, *searchoutput.ListSearchOutput)
	}{
		{
			name:  "success - returns specialists with pagination built by command",
			input: searchDTOFactory(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(searchResultFactory(), nil).Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				require.NotNil(t, output)
				assert.Len(t, output.Specialists, 1)
				assert.Equal(t, "Dr. João Silva", output.Specialists[0].Name)
				require.NotNil(t, output.CursorOutput)
				assert.False(t, output.CursorOutput.HasNextPage)
				assert.False(t, output.CursorOutput.HasPreviousPage)
				assert.Nil(t, output.CursorOutput.NextCursor)
				assert.Nil(t, output.CursorOutput.PreviousCursor)
				assert.Equal(t, 1, output.CursorOutput.TotalItemsInPage)
			},
		},
		{
			name:  "success - builds next cursor when repository signals has next page",
			input: searchDTOFactory(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				result := searchResultFactory(func(r *searchoutput.SearchResult) {
					r.HasNextPage = true
					r.LastSortValues = []any{4.8, "2025-01-01T00:00:00Z", "last-id"}
				})
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(result, nil).Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				require.NotNil(t, output)
				assert.True(t, output.CursorOutput.HasNextPage)
				assert.NotNil(t, output.CursorOutput.NextCursor)
				assert.NotEmpty(t, *output.CursorOutput.NextCursor)
				assert.False(t, output.CursorOutput.HasPreviousPage)
				assert.Nil(t, output.CursorOutput.PreviousCursor)
			},
		},
		{
			name: "success - builds previous cursor when navigating beyond first page",
			input: func() *SearchSpecialistsDTO {
				encodedCursor := cursor.EncodeCursorMultiSort([]any{4.5, "2025-01-01T00:00:00Z", "some-id"})
				pagination, _ := cursor.NewCursorPaginationInput(&encodedCursor, 10, cursor.DirectionNext)
				searchTerm := "cardiologia"
				return &SearchSpecialistsDTO{
					SearchTerm: &searchTerm,
					Filters:    nil,
					Sort:       nil,
					Pagination: pagination,
				}
			}(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				result := searchResultFactory(func(r *searchoutput.SearchResult) {
					r.HasNextPage = false
					r.FirstSortValues = []any{4.2, "2025-01-02T00:00:00Z", "first-id"}
					r.LastSortValues = []any{3.8, "2025-01-03T00:00:00Z", "last-id"}
				})
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(result, nil).Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				require.NotNil(t, output)
				assert.False(t, output.CursorOutput.HasNextPage)
				assert.Nil(t, output.CursorOutput.NextCursor)
				assert.True(t, output.CursorOutput.HasPreviousPage)
				assert.NotNil(t, output.CursorOutput.PreviousCursor)
				assert.NotEmpty(t, *output.CursorOutput.PreviousCursor)
			},
		},
		{
			name:  "success - returns empty output when no specialists match",
			input: searchDTOFactory(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(emptySearchResult(), nil).Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				require.NotNil(t, output)
				assert.Empty(t, output.Specialists)
				assert.True(t, output.IsEmpty())
				assert.False(t, output.CursorOutput.HasNextPage)
				assert.Nil(t, output.CursorOutput.NextCursor)
				assert.Nil(t, output.CursorOutput.PreviousCursor)
				assert.Equal(t, 0, output.CursorOutput.TotalItemsInPage)
			},
		},
		{
			name:  "success - does not generate next cursor when hasNext but no sort values",
			input: searchDTOFactory(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				result := searchResultFactory(func(r *searchoutput.SearchResult) {
					r.HasNextPage = true
					r.LastSortValues = nil
				})
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(result, nil).Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				require.NotNil(t, output)
				assert.True(t, output.CursorOutput.HasNextPage)
				assert.Nil(t, output.CursorOutput.NextCursor)
			},
		},
		{
			name:  "failure - returns ErrInvalidSearchInput when dto is nil",
			input: nil,
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				logger.EXPECT().Error(gomock.Any(), ErrInvalidSearchInputMessage).Times(1)
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrInvalidSearchInput,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				assert.Nil(t, output)
			},
		},
		{
			name: "failure - returns ErrInvalidSearchInput when domain validation fails",
			input: func() *SearchSpecialistsDTO {
				pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)
				return &SearchSpecialistsDTO{
					SearchTerm: nil,
					Filters:    nil,
					Sort:       nil,
					Pagination: pagination,
				}
			}(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				logger.EXPECT().Error(gomock.Any(), ErrInvalidSearchInputMessage, gomock.Any()).Times(1)
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrInvalidSearchInput,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				assert.Nil(t, output)
			},
		},
		{
			name:  "failure - returns ErrSearchExecution when repository returns infrastructure error",
			input: searchDTOFactory(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, errors.New("connection refused")).Times(1)
				logger.EXPECT().Error(gomock.Any(), ErrSearchExecutionMessage, gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: ErrSearchExecution,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				assert.Nil(t, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSpecialistSearchRepositoryInterface(ctrl)
			mockLogger := loggerMocks.NewMockLogger(ctrl)

			tt.setupMocks(mockRepo, mockLogger)

			command := NewSearchSpecialistsCommand(mockRepo, mockLogger)

			output, err := command.Execute(context.Background(), tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, output)
		})
	}
}
