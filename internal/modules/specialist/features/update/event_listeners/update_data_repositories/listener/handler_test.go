package listener

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/event_listeners/update_data_repositories/listener/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func specialistFactory(overrides ...func(*domain.Specialist)) *domain.Specialist {
	now := time.Now().UTC()
	s := &domain.Specialist{
		ID:            "specialist-123",
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

func payloadFactory(overrides ...func(*UpdateDataRepositoriesEventPayload)) UpdateDataRepositoriesEventPayload {
	p := UpdateDataRepositoriesEventPayload{
		ID: "specialist-123",
	}
	for _, o := range overrides {
		o(&p)
	}
	return p
}

func makeEvent(payload UpdateDataRepositoriesEventPayload) event.Event {
	data, _ := json.Marshal(payload)
	return event.NewEvent("specialist.updated", data)
}

func zeroDelayRetryConfig() event.RetryConfig {
	return event.RetryConfig{
		MaxRetries: 3,
		Delay:      0,
	}
}

func TestUpdateDataRepositoriesHandler_Handle(t *testing.T) {
	tests := []struct {
		name        string
		event       event.Event
		setupMocks  func(*mocks.MockSourceRepository, *mocks.MockDataRepository)
		repoCount   int
		expectError bool
		expectedErr error
	}{
		{
			name:      "success - fetches specialist and updates single data repository",
			event:     makeEvent(payloadFactory()),
			repoCount: 1,
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository) {
				specialist := specialistFactory()

				mockSource.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockData.EXPECT().Update(gomock.Any(), specialist).Return(nil).Times(1)
			},
			expectError: false,
		},
		{
			name:  "failure - returns error when payload is invalid JSON",
			event: event.NewEvent("specialist.updated", []byte("invalid-json")),
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository) {
				mockSource.EXPECT().FindByID(gomock.Any(), gomock.Any()).Times(0)
				mockData.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:      "failure - returns ErrSpecialistNotFound when source repository FindByID fails",
			event:     makeEvent(payloadFactory()),
			repoCount: 1,
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository) {
				mockSource.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(nil, errors.New("not found")).Times(1)
				mockData.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrSpecialistNotFound,
		},
		{
			name:      "failure - returns ErrUpdateDataRepositories when repository update fails after retries",
			event:     makeEvent(payloadFactory()),
			repoCount: 1,
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository) {
				specialist := specialistFactory()

				mockSource.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockData.EXPECT().Update(gomock.Any(), specialist).Return(errors.New("es unavailable")).Times(3)
			},
			expectError: true,
			expectedErr: ErrUpdateDataRepositories,
		},
		{
			name:      "success - succeeds when all repositories update successfully after initial retry failures",
			event:     makeEvent(payloadFactory()),
			repoCount: 1,
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository) {
				specialist := specialistFactory()

				mockSource.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)

				callCount := 0
				mockData.EXPECT().Update(gomock.Any(), specialist).DoAndReturn(func(ctx context.Context, s *domain.Specialist) error {
					callCount++
					if callCount < 3 {
						return errors.New("temporary failure")
					}
					return nil
				}).Times(3)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSource := mocks.NewMockSourceRepository(ctrl)
			mockData := mocks.NewMockDataRepository(ctrl)

			tt.setupMocks(mockSource, mockData)

			dataRepos := []DataRepository{}
			for range tt.repoCount {
				dataRepos = append(dataRepos, mockData)
			}

			handler := NewUpdateDataRepositoriesHandler(
				mockSource,
				dataRepos,
			).WithRetryConfig(zeroDelayRetryConfig())

			err := handler.Handle(context.Background(), tt.event)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
