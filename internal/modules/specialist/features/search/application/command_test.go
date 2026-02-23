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
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application/mocks"
)

func searchInputFactory(overrides ...func(*searchinput.ListSearchInput)) *searchinput.ListSearchInput {
	searchTerm := "cardiologia"
	pagination, _ := cursor.NewCursorPaginationInput(nil, 10, cursor.DirectionNext)

	input, _ := searchinput.NewListSearchInput(
		&searchTerm,
		nil,
		nil,
		pagination,
	)

	for _, override := range overrides {
		override(input)
	}

	return input
}

func searchOutputFactory(overrides ...func(*searchoutput.ListSearchOutput)) *searchoutput.ListSearchOutput {
	now := time.Now()
	specialists := []*domain.Specialist{
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
	}

	cursorOutput := cursor.NewCursorPaginationOutput(nil, nil, false, false, len(specialists))

	output := searchoutput.NewListSearchOutput(specialists, cursorOutput)

	for _, override := range overrides {
		override(output)
	}

	return output
}

func emptySearchOutputFactory() *searchoutput.ListSearchOutput {
	cursorOutput := cursor.NewCursorPaginationOutput(nil, nil, false, false, 0)
	return searchoutput.NewListSearchOutput([]*domain.Specialist{}, cursorOutput)
}

func TestSearchSpecialistsCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		input          *searchinput.ListSearchInput
		setupMocks     func(*mocks.MockSpecialistSearchRepositoryInterface, *loggerMocks.MockLogger)
		expectError    bool
		expectedErr    error
		validateResult func(*testing.T, *searchoutput.ListSearchOutput)
	}{
		{
			name:  "success - returns specialists when search finds results",
			input: searchInputFactory(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				logger.EXPECT().Info(gomock.Any(), StartingSearchMessage).Times(1)
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(searchOutputFactory(), nil).Times(1)
				logger.EXPECT().Info(gomock.Any(), SearchCompletedMessage, gomock.Any()).Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				require.NotNil(t, output)
				assert.Len(t, output.Specialists, 1)
				assert.Equal(t, "Dr. João Silva", output.Specialists[0].Name)
				assert.NotNil(t, output.CursorOutput)
			},
		},
		{
			name:  "success - returns empty output when no specialists match",
			input: searchInputFactory(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				logger.EXPECT().Info(gomock.Any(), StartingSearchMessage).Times(1)
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(emptySearchOutputFactory(), nil).Times(1)
				logger.EXPECT().Info(gomock.Any(), SearchNoResultsMessage).Times(1)
			},
			expectError: false,
			expectedErr: nil,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				require.NotNil(t, output)
				assert.Empty(t, output.Specialists)
				assert.True(t, output.IsEmpty())
			},
		},
		{
			name:  "failure - returns ErrInvalidSearchInput when input is nil",
			input: nil,
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				logger.EXPECT().Info(gomock.Any(), StartingSearchMessage).Times(1)
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
			name:  "failure - returns ErrSearchExecution when repository returns infrastructure error",
			input: searchInputFactory(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				logger.EXPECT().Info(gomock.Any(), StartingSearchMessage).Times(1)
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, errors.New("connection refused")).Times(1)
				logger.EXPECT().Error(gomock.Any(), ErrSearchExecutionMessage, gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: ErrSearchExecution,
			validateResult: func(t *testing.T, output *searchoutput.ListSearchOutput) {
				assert.Nil(t, output)
			},
		},
		{
			name:  "failure - returns ErrInvalidSearchInput when repository returns domain error",
			input: searchInputFactory(),
			setupMocks: func(repo *mocks.MockSpecialistSearchRepositoryInterface, logger *loggerMocks.MockLogger) {
				logger.EXPECT().Info(gomock.Any(), StartingSearchMessage).Times(1)
				repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, search.ErrEmptySearchCriteria).Times(1)
				logger.EXPECT().Error(gomock.Any(), ErrSearchExecutionMessage, gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: ErrInvalidSearchInput,
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
