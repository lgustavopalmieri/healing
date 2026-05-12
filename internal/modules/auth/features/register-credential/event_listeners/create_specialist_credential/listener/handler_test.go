package listener_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/create_specialist_credential/listener"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/create_specialist_credential/listener/mocks"
	authevents "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/events"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

const (
	testSpecialistID    = "specialist-abc-123"
	testSpecialistEmail = "specialist@healing.com"
	testLicense         = "CRM-12345"
	testSpecialty       = "cardiology"
	testSetPasswordJWT  = "eyJhbGci.signed.token"
	testSetPasswordJTI  = "jti-xyz-789"
)

var (
	errRepoFailure  = errors.New("database connection refused")
	errSaveFailure  = errors.New("unique violation on credential save")
	errTokenFailure = errors.New("redis set failed")
)

func specialistCreatedPayload(overrides ...func(*listener.SpecialistCreatedPayload)) []byte {
	p := listener.SpecialistCreatedPayload{
		ID:            testSpecialistID,
		Email:         testSpecialistEmail,
		LicenseNumber: testLicense,
		Specialty:     testSpecialty,
	}
	for _, o := range overrides {
		o(&p)
	}
	raw, _ := json.Marshal(p)
	return raw
}

func TestCreateSpecialistCredentialHandler_Handle(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		setupMocks func(
			credentialRepository *mocks.MockCredentialRepository,
			setPasswordTokenGenerator *mocks.MockSetPasswordTokenGenerator,
			eventPublisher *mocks.MockEventDispatcher,
		)
		expectError    bool
		errMsgContains string
	}{
		{
			name:    "happy path - cria credential pending, gera token, publica auth.credential.pending",
			payload: specialistCreatedPayload(),
			setupMocks: func(
				credentialRepository *mocks.MockCredentialRepository,
				setPasswordTokenGenerator *mocks.MockSetPasswordTokenGenerator,
				eventPublisher *mocks.MockEventDispatcher,
			) {
				credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testSpecialistEmail, provider.Password, role.Specialist).
					Times(1).
					Return(nil, nil)

				credentialRepository.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, c *credential.Credential) error {
						assert.Equal(t, testSpecialistID, c.SubjectID)
						assert.Equal(t, role.Specialist, c.Role)
						assert.Equal(t, provider.Password, c.Provider)
						assert.Equal(t, testSpecialistEmail, c.Email)
						assert.Equal(t, credential.StatusPending, c.Status)
						assert.NotEmpty(t, c.ID)
						return nil
					})

				setPasswordTokenGenerator.EXPECT().
					Generate(gomock.Any(), testSpecialistID).
					Times(1).
					Return(testSetPasswordJWT, testSetPasswordJTI, nil)

				eventPublisher.EXPECT().
					Dispatch(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, evt event.Event) error {
						assert.Equal(t, authevents.AuthCredentialPending, evt.Name)
						payload, ok := evt.Payload.(map[string]any)
						require.True(t, ok)
						assert.Equal(t, testSpecialistID, payload["subject_id"])
						assert.Equal(t, role.Specialist.String(), payload["role"])
						assert.Equal(t, testSpecialistEmail, payload["email"])
						assert.Equal(t, testSetPasswordJWT, payload["set_password_token"])
						return nil
					})
			},
		},
		{
			name:    "happy path - idempotencia: credential ja existente retorna nil sem Save/Generate/Dispatch",
			payload: specialistCreatedPayload(),
			setupMocks: func(
				credentialRepository *mocks.MockCredentialRepository,
				setPasswordTokenGenerator *mocks.MockSetPasswordTokenGenerator,
				eventPublisher *mocks.MockEventDispatcher,
			) {
				credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testSpecialistEmail, provider.Password, role.Specialist).
					Times(1).
					Return(&credential.Credential{ID: "existing-cred"}, nil)

				credentialRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				setPasswordTokenGenerator.EXPECT().Generate(gomock.Any(), gomock.Any()).Times(0)
				eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name:    "failure - payload nao-bytes retorna ErrInvalidEventPayload sem tocar em colaboradores",
			payload: "not a byte slice",
			setupMocks: func(
				credentialRepository *mocks.MockCredentialRepository,
				setPasswordTokenGenerator *mocks.MockSetPasswordTokenGenerator,
				eventPublisher *mocks.MockEventDispatcher,
			) {
				credentialRepository.EXPECT().FindByEmailProviderRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				credentialRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				setPasswordTokenGenerator.EXPECT().Generate(gomock.Any(), gomock.Any()).Times(0)
				eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError:    true,
			errMsgContains: listener.ErrInvalidEventPayloadMessage,
		},
		{
			name:    "failure - payload JSON malformado retorna erro envelopando ErrUnmarshalEventPayloadMessage",
			payload: []byte("{not-valid-json"),
			setupMocks: func(
				credentialRepository *mocks.MockCredentialRepository,
				setPasswordTokenGenerator *mocks.MockSetPasswordTokenGenerator,
				eventPublisher *mocks.MockEventDispatcher,
			) {
				credentialRepository.EXPECT().FindByEmailProviderRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				credentialRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				setPasswordTokenGenerator.EXPECT().Generate(gomock.Any(), gomock.Any()).Times(0)
				eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError:    true,
			errMsgContains: listener.ErrUnmarshalEventPayloadMessage,
		},
		{
			name:    "failure - FindByEmailProviderRole retorna erro envelopado com ErrFindCredentialMessage",
			payload: specialistCreatedPayload(),
			setupMocks: func(
				credentialRepository *mocks.MockCredentialRepository,
				setPasswordTokenGenerator *mocks.MockSetPasswordTokenGenerator,
				eventPublisher *mocks.MockEventDispatcher,
			) {
				credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testSpecialistEmail, provider.Password, role.Specialist).
					Times(1).
					Return(nil, errRepoFailure)
				credentialRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				setPasswordTokenGenerator.EXPECT().Generate(gomock.Any(), gomock.Any()).Times(0)
				eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError:    true,
			errMsgContains: listener.ErrFindCredentialMessage,
		},
		{
			name:    "failure - Save retorna erro envelopado com ErrSaveCredentialMessage",
			payload: specialistCreatedPayload(),
			setupMocks: func(
				credentialRepository *mocks.MockCredentialRepository,
				setPasswordTokenGenerator *mocks.MockSetPasswordTokenGenerator,
				eventPublisher *mocks.MockEventDispatcher,
			) {
				credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testSpecialistEmail, provider.Password, role.Specialist).
					Times(1).
					Return(nil, nil)
				credentialRepository.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(1).
					Return(errSaveFailure)
				setPasswordTokenGenerator.EXPECT().Generate(gomock.Any(), gomock.Any()).Times(0)
				eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError:    true,
			errMsgContains: listener.ErrSaveCredentialMessage,
		},
		{
			name:    "failure - Generate retorna erro envelopado com ErrGenerateSetPasswordMessage",
			payload: specialistCreatedPayload(),
			setupMocks: func(
				credentialRepository *mocks.MockCredentialRepository,
				setPasswordTokenGenerator *mocks.MockSetPasswordTokenGenerator,
				eventPublisher *mocks.MockEventDispatcher,
			) {
				credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testSpecialistEmail, provider.Password, role.Specialist).
					Times(1).
					Return(nil, nil)
				credentialRepository.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
				setPasswordTokenGenerator.EXPECT().
					Generate(gomock.Any(), testSpecialistID).
					Times(1).
					Return("", "", errTokenFailure)
				eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError:    true,
			errMsgContains: listener.ErrGenerateSetPasswordMessage,
		},
		{
			name:    "happy path - Dispatch retornando erro nao propaga (publish fire-and-forget seguindo padrao do projeto)",
			payload: specialistCreatedPayload(),
			setupMocks: func(
				credentialRepository *mocks.MockCredentialRepository,
				setPasswordTokenGenerator *mocks.MockSetPasswordTokenGenerator,
				eventPublisher *mocks.MockEventDispatcher,
			) {
				credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testSpecialistEmail, provider.Password, role.Specialist).
					Times(1).
					Return(nil, nil)
				credentialRepository.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
				setPasswordTokenGenerator.EXPECT().
					Generate(gomock.Any(), testSpecialistID).
					Times(1).
					Return(testSetPasswordJWT, testSetPasswordJTI, nil)
				eventPublisher.EXPECT().
					Dispatch(gomock.Any(), gomock.Any()).
					Times(1).
					Return(errors.New("sqs down"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			credentialRepository := mocks.NewMockCredentialRepository(ctrl)
			setPasswordTokenGenerator := mocks.NewMockSetPasswordTokenGenerator(ctrl)
			eventPublisher := mocks.NewMockEventDispatcher(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(credentialRepository, setPasswordTokenGenerator, eventPublisher)
			}

			handler := listener.NewCreateSpecialistCredentialHandler(
				credentialRepository,
				setPasswordTokenGenerator,
				eventPublisher,
			)

			evt := event.NewEvent(listener.SpecialistCreatedEventName, tt.payload)
			err := handler.Handle(context.Background(), evt)

			if tt.expectError {
				require.Error(t, err)
				if tt.errMsgContains != "" {
					assert.Contains(t, err.Error(), tt.errMsgContains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}
