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
		setupMocks  func(*mocks.MockSourceRepository, *mocks.MockDataRepository, *mocks.MockTracer, *mocks.MockLogger, *mocks.MockSpan)
		repoCount   int
		expectError bool
		expectedErr error
	}{
		{
			name:      "success - fetches specialist and updates single data repository",
			event:     makeEvent(payloadFactory()),
			repoCount: 1,
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan) {
				ctx := context.Background()
				specialist := specialistFactory()

				mockTracer.EXPECT().Start(gomock.Any(), UpdateDataRepositoriesSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockSource.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockLogger.EXPECT().Info(gomock.Any(), StartingDataRepositoriesUpdateMessage, gomock.Any()).Times(1)

				mockData.EXPECT().Update(gomock.Any(), specialist).Return(nil).Times(1)
				mockLogger.EXPECT().Info(gomock.Any(), RepositoryUpdateSucceededMessage, gomock.Any()).Times(1)

				mockData.EXPECT().PublishDLQ(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				mockLogger.EXPECT().Info(gomock.Any(), DataRepositoriesUpdatedSuccessMessage, gomock.Any()).Times(1)
			},
			expectError: false,
		},
		{
			name:  "failure - returns error when payload is invalid JSON",
			event: event.NewEvent("specialist.updated", []byte("invalid-json")),
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan) {
				ctx := context.Background()
				mockTracer.EXPECT().Start(gomock.Any(), UpdateDataRepositoriesSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockSource.EXPECT().FindByID(gomock.Any(), gomock.Any()).Times(0)
				mockData.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
				mockData.EXPECT().PublishDLQ(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:      "failure - returns ErrSpecialistNotFound when source repository FindByID fails",
			event:     makeEvent(payloadFactory()),
			repoCount: 1,
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan) {
				ctx := context.Background()

				mockTracer.EXPECT().Start(gomock.Any(), UpdateDataRepositoriesSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockSource.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(nil, errors.New("not found")).Times(1)
				mockSpan.EXPECT().RecordError(gomock.Any()).Times(1)
				mockLogger.EXPECT().Error(gomock.Any(), ErrSpecialistNotFoundMessage, gomock.Any(), gomock.Any()).Times(1)

				mockData.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
				mockData.EXPECT().PublishDLQ(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrSpecialistNotFound,
		},
		{
			name:      "failure - returns ErrUpdateDataRepositories when repository update fails after retries and publishes DLQ",
			event:     makeEvent(payloadFactory()),
			repoCount: 1,
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan) {
				ctx := context.Background()
				specialist := specialistFactory()

				mockTracer.EXPECT().Start(gomock.Any(), UpdateDataRepositoriesSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockSource.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockLogger.EXPECT().Info(gomock.Any(), StartingDataRepositoriesUpdateMessage, gomock.Any()).Times(1)

				mockData.EXPECT().Update(gomock.Any(), specialist).Return(errors.New("es unavailable")).Times(3)

				mockSpan.EXPECT().RecordError(gomock.Any()).Times(1)
				mockLogger.EXPECT().Error(gomock.Any(), RepositoryUpdateFailedMessage, gomock.Any(), gomock.Any()).Times(1)

				mockData.EXPECT().PublishDLQ(gomock.Any(), specialist, gomock.Any()).Return(nil).Times(1)
			},
			expectError: true,
			expectedErr: ErrUpdateDataRepositories,
		},
		{
			name:      "failure - returns ErrUpdateDataRepositories when repository update fails and DLQ publish also fails",
			event:     makeEvent(payloadFactory()),
			repoCount: 1,
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan) {
				ctx := context.Background()
				specialist := specialistFactory()

				mockTracer.EXPECT().Start(gomock.Any(), UpdateDataRepositoriesSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockSource.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockLogger.EXPECT().Info(gomock.Any(), StartingDataRepositoriesUpdateMessage, gomock.Any()).Times(1)

				mockData.EXPECT().Update(gomock.Any(), specialist).Return(errors.New("es unavailable")).Times(3)

				mockSpan.EXPECT().RecordError(gomock.Any()).Times(2)
				mockLogger.EXPECT().Error(gomock.Any(), RepositoryUpdateFailedMessage, gomock.Any(), gomock.Any()).Times(1)

				mockData.EXPECT().PublishDLQ(gomock.Any(), specialist, gomock.Any()).Return(errors.New("kafka down")).Times(1)
				mockLogger.EXPECT().Error(gomock.Any(), DLQPublishFailedMessage, gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: ErrUpdateDataRepositories,
		},
		{
			name:      "success - succeeds when all repositories update successfully after initial retry failures",
			event:     makeEvent(payloadFactory()),
			repoCount: 1,
			setupMocks: func(mockSource *mocks.MockSourceRepository, mockData *mocks.MockDataRepository, mockTracer *mocks.MockTracer, mockLogger *mocks.MockLogger, mockSpan *mocks.MockSpan) {
				ctx := context.Background()
				specialist := specialistFactory()

				mockTracer.EXPECT().Start(gomock.Any(), UpdateDataRepositoriesSpanName).Return(ctx, mockSpan).Times(1)
				mockSpan.EXPECT().End().Times(1)

				mockSource.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockLogger.EXPECT().Info(gomock.Any(), StartingDataRepositoriesUpdateMessage, gomock.Any()).Times(1)

				callCount := 0
				mockData.EXPECT().Update(gomock.Any(), specialist).DoAndReturn(func(ctx context.Context, s *domain.Specialist) error {
					callCount++
					if callCount < 3 {
						return errors.New("temporary failure")
					}
					return nil
				}).Times(3)

				mockData.EXPECT().PublishDLQ(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				mockLogger.EXPECT().Info(gomock.Any(), RepositoryUpdateSucceededMessage, gomock.Any()).Times(1)
				mockLogger.EXPECT().Info(gomock.Any(), DataRepositoriesUpdatedSuccessMessage, gomock.Any()).Times(1)
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
			mockTracer := mocks.NewMockTracer(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)
			mockSpan := mocks.NewMockSpan(ctrl)

			tt.setupMocks(mockSource, mockData, mockTracer, mockLogger, mockSpan)

			dataRepos := []DataRepository{}
			for range tt.repoCount {
				dataRepos = append(dataRepos, mockData)
			}

			handler := NewUpdateDataRepositoriesHandler(
				mockSource,
				dataRepos,
				mockTracer,
				mockLogger,
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
