package listener

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	authorizelicense "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/authorize_license"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/listener/mocks"
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
		Rating:        0.0,
		Status:        domain.StatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, o := range overrides {
		o(s)
	}
	return s
}

func payloadFactory(overrides ...func(*ValidateLicenseEventPayload)) ValidateLicenseEventPayload {
	p := ValidateLicenseEventPayload{
		ID:            "specialist-123",
		Email:         "joao@example.com",
		LicenseNumber: "CRM123456",
		Specialty:     "Cardiologia",
	}
	for _, o := range overrides {
		o(&p)
	}
	return p
}

func makeEvent(payload ValidateLicenseEventPayload) event.Event {
	data, _ := json.Marshal(payload)
	return event.NewEvent("specialist.created", data)
}

func TestValidateLicenseHandler_Handle(t *testing.T) {
	tests := []struct {
		name        string
		event       event.Event
		setupMocks  func(*mocks.MockValidateLicenseRepositoryInterface, *mocks.MockLicenseGatewayInterface, *mocks.MockEventDispatcher)
		expectError bool
		expectedErr error
	}{
		{
			name:  "success - validates license and updates status to authorized_license",
			event: makeEvent(payloadFactory()),
			setupMocks: func(mockRepo *mocks.MockValidateLicenseRepositoryInterface, mockGateway *mocks.MockLicenseGatewayInterface, mockEvent *mocks.MockEventDispatcher) {
				specialist := specialistFactory()

				mockRepo.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockGateway.EXPECT().Validate(gomock.Any(), "CRM123456").Return(true, nil).Times(1)
				mockRepo.EXPECT().UpdateStatus(gomock.Any(), "specialist-123", domain.StatusAuthorizedLicense).Return(nil).Times(1)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectError: false,
		},
		{
			name:  "failure - returns error when payload is invalid JSON",
			event: event.NewEvent("specialist.created", []byte("invalid-json")),
			setupMocks: func(mockRepo *mocks.MockValidateLicenseRepositoryInterface, mockGateway *mocks.MockLicenseGatewayInterface, mockEvent *mocks.MockEventDispatcher) {
				mockRepo.EXPECT().FindByID(gomock.Any(), gomock.Any()).Times(0)
				mockGateway.EXPECT().Validate(gomock.Any(), gomock.Any()).Times(0)
				mockRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:  "failure - returns ErrSpecialistNotFound when repository FindByID fails",
			event: makeEvent(payloadFactory()),
			setupMocks: func(mockRepo *mocks.MockValidateLicenseRepositoryInterface, mockGateway *mocks.MockLicenseGatewayInterface, mockEvent *mocks.MockEventDispatcher) {
				mockRepo.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(nil, errors.New("not found")).Times(1)

				mockGateway.EXPECT().Validate(gomock.Any(), gomock.Any()).Times(0)
				mockRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrSpecialistNotFound,
		},
		{
			name:  "failure - returns ErrLicenseValidation when gateway returns error",
			event: makeEvent(payloadFactory()),
			setupMocks: func(mockRepo *mocks.MockValidateLicenseRepositoryInterface, mockGateway *mocks.MockLicenseGatewayInterface, mockEvent *mocks.MockEventDispatcher) {
				specialist := specialistFactory()

				mockRepo.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockGateway.EXPECT().Validate(gomock.Any(), "CRM123456").Return(false, errors.New("service unavailable")).Times(1)

				mockRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrLicenseValidation,
		},
		{
			name:  "failure - returns ErrInvalidLicense when gateway returns false",
			event: makeEvent(payloadFactory()),
			setupMocks: func(mockRepo *mocks.MockValidateLicenseRepositoryInterface, mockGateway *mocks.MockLicenseGatewayInterface, mockEvent *mocks.MockEventDispatcher) {
				specialist := specialistFactory()

				mockRepo.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockGateway.EXPECT().Validate(gomock.Any(), "CRM123456").Return(false, nil).Times(1)

				mockRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrInvalidLicense,
		},
		{
			name:  "failure - returns ErrInvalidStatusTransition when specialist is not pending",
			event: makeEvent(payloadFactory()),
			setupMocks: func(mockRepo *mocks.MockValidateLicenseRepositoryInterface, mockGateway *mocks.MockLicenseGatewayInterface, mockEvent *mocks.MockEventDispatcher) {
				specialist := specialistFactory(func(s *domain.Specialist) {
					s.Status = domain.StatusActive
				})

				mockRepo.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockGateway.EXPECT().Validate(gomock.Any(), "CRM123456").Return(true, nil).Times(1)

				mockRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: authorizelicense.ErrInvalidStatusTransition,
		},
		{
			name:  "failure - returns ErrUpdateStatus when repository UpdateStatus fails",
			event: makeEvent(payloadFactory()),
			setupMocks: func(mockRepo *mocks.MockValidateLicenseRepositoryInterface, mockGateway *mocks.MockLicenseGatewayInterface, mockEvent *mocks.MockEventDispatcher) {
				specialist := specialistFactory()

				mockRepo.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockGateway.EXPECT().Validate(gomock.Any(), "CRM123456").Return(true, nil).Times(1)
				mockRepo.EXPECT().UpdateStatus(gomock.Any(), "specialist-123", domain.StatusAuthorizedLicense).Return(errors.New("db error")).Times(1)

				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: ErrUpdateStatus,
		},
		{
			name:  "success - still succeeds when event publish fails",
			event: makeEvent(payloadFactory()),
			setupMocks: func(mockRepo *mocks.MockValidateLicenseRepositoryInterface, mockGateway *mocks.MockLicenseGatewayInterface, mockEvent *mocks.MockEventDispatcher) {
				specialist := specialistFactory()

				mockRepo.EXPECT().FindByID(gomock.Any(), "specialist-123").Return(specialist, nil).Times(1)
				mockGateway.EXPECT().Validate(gomock.Any(), "CRM123456").Return(true, nil).Times(1)
				mockRepo.EXPECT().UpdateStatus(gomock.Any(), "specialist-123", domain.StatusAuthorizedLicense).Return(nil).Times(1)
				mockEvent.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Return(errors.New("kafka unavailable")).Times(1)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockValidateLicenseRepositoryInterface(ctrl)
			mockGateway := mocks.NewMockLicenseGatewayInterface(ctrl)
			mockEvent := mocks.NewMockEventDispatcher(ctrl)

			tt.setupMocks(mockRepo, mockGateway, mockEvent)

			handler := NewValidateLicenseHandler(mockRepo, mockGateway, mockEvent)

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
