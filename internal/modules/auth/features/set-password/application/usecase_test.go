package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/password"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/application/mocks"
	authevents "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/events"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

const (
	testRawToken      = "eyJ.set.password.token"
	testJTI           = "jti-set-password-123"
	testSubjectID     = "subject-uuid-abc"
	testEmail         = "specialist@healing.com"
	testPassword      = "abc12345"
	testWeakNoDigit   = "abcdefgh"
	testWeakShort     = "ab12"
	testDeviceInfo    = "web"
	testIPAddress     = "1.2.3.4"
	testUserAgent     = "go-test"
	testAccessJWT     = "access.jwt.value"
	testAccessJTI     = "access-jti-xyz"
	testRefreshOpaque = "opaque-refresh-token-base64"
)

var (
	errValidator    = errors.New("invalid token")
	errRedisDown    = errors.New("redis unreachable")
	errPostgresDown = errors.New("postgres unreachable")
	errIssuer       = errors.New("signer failed")
	errAudit        = errors.New("audit insert failed")
	errEvent        = errors.New("sns publish failed")
	errRefreshCache = errors.New("redis set failed")
)

type useCaseMocks struct {
	tokenValidator           *mocks.MockSetPasswordTokenValidator
	singleUseTokenRepository *mocks.MockSingleUseTokenRepository
	credentialRepository     *mocks.MockCredentialRepository
	accessTokenIssuer        *mocks.MockAccessTokenIssuer
	refreshTokenRepository   *mocks.MockRefreshTokenRepository
	auditRepository          *mocks.MockAuditRepository
	eventPublisher           *mocks.MockEventDispatcher
	logger                   *mocks.MockLogger
}

func newUseCaseMocks(ctrl *gomock.Controller) *useCaseMocks {
	return &useCaseMocks{
		tokenValidator:           mocks.NewMockSetPasswordTokenValidator(ctrl),
		singleUseTokenRepository: mocks.NewMockSingleUseTokenRepository(ctrl),
		credentialRepository:     mocks.NewMockCredentialRepository(ctrl),
		accessTokenIssuer:        mocks.NewMockAccessTokenIssuer(ctrl),
		refreshTokenRepository:   mocks.NewMockRefreshTokenRepository(ctrl),
		auditRepository:          mocks.NewMockAuditRepository(ctrl),
		eventPublisher:           mocks.NewMockEventDispatcher(ctrl),
		logger:                   mocks.NewMockLogger(ctrl),
	}
}

func (m *useCaseMocks) build() *application.SetPasswordUseCase {
	return application.NewSetPasswordUseCase(application.SetPasswordUseCaseDependencies{
		TokenValidator:           m.tokenValidator,
		SingleUseTokenRepository: m.singleUseTokenRepository,
		CredentialRepository:     m.credentialRepository,
		AccessTokenIssuer:        m.accessTokenIssuer,
		RefreshTokenRepository:   m.refreshTokenRepository,
		AuditRepository:          m.auditRepository,
		EventPublisher:           m.eventPublisher,
		Logger:                   m.logger,
		PasswordMinLength:        8,
		BcryptCost:               4,
	})
}

func validatedTokenFactory(overrides ...func(*application.ValidatedSetPasswordToken)) *application.ValidatedSetPasswordToken {
	v := &application.ValidatedSetPasswordToken{
		SubjectID: testSubjectID,
		Role:      role.Specialist,
		JTI:       testJTI,
	}
	for _, o := range overrides {
		o(v)
	}
	return v
}

func pendingCredentialFactory(overrides ...func(*credential.Credential)) *credential.Credential {
	cred := credential.NewCredential(credential.NewCredentialInput{
		SubjectID: testSubjectID,
		Role:      role.Specialist,
		Provider:  provider.Password,
		Email:     testEmail,
	})
	for _, o := range overrides {
		o(cred)
	}
	return cred
}

func issuedTokenPairFactory() *tokenpair.TokenPair {
	now := time.Now()
	return &tokenpair.TokenPair{
		AccessToken:      testAccessJWT,
		AccessJTI:        testAccessJTI,
		AccessExpiresAt:  now.Add(1 * time.Hour),
		RefreshToken:     testRefreshOpaque,
		RefreshExpiresAt: now.Add(168 * time.Hour),
	}
}

func inputFactory(overrides ...func(*application.SetPasswordDTO)) application.SetPasswordDTO {
	in := application.SetPasswordDTO{
		Token:      testRawToken,
		Password:   testPassword,
		DeviceInfo: testDeviceInfo,
		IPAddress:  testIPAddress,
		UserAgent:  testUserAgent,
	}
	for _, o := range overrides {
		o(&in)
	}
	return in
}

func expectLoggerAnyError(logger *mocks.MockLogger, times int) {
	logger.EXPECT().
		Error(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(times)
}

func TestSetPasswordUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		input          application.SetPasswordDTO
		setupMocks     func(m *useCaseMocks)
		expectError    bool
		expectedErr    error
		validateResult func(t *testing.T, result *application.SetPasswordResult)
	}{
		{
			name:  "happy path - fluxo completo: ativa credential, emite tokens, persiste, cacheia refresh, publica audit+event em paralelo",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				validated := validatedTokenFactory()
				cred := pendingCredentialFactory()
				issued := issuedTokenPairFactory()

				m.tokenValidator.EXPECT().
					Validate(gomock.Any(), testRawToken).
					Times(1).
					Return(validated, nil)
				m.singleUseTokenRepository.EXPECT().
					Consume(gomock.Any(), testJTI).
					Times(1).
					Return(true, nil)
				m.credentialRepository.EXPECT().
					FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).
					Times(1).
					Return(cred, nil)
				m.accessTokenIssuer.EXPECT().
					IssueAccessAndRefresh(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, c *credential.Credential) (*tokenpair.TokenPair, error) {
						assert.Equal(t, credential.StatusActive, c.Status)
						assert.False(t, c.PasswordHash.IsEmpty())
						return issued, nil
					})
				m.credentialRepository.EXPECT().
					UpdateWithSessionInTransaction(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, c *credential.Credential, s *session.Session) error {
						assert.Equal(t, credential.StatusActive, c.Status)
						assert.Equal(t, testSubjectID, s.SubjectID)
						assert.NotEmpty(t, s.RefreshTokenHash)
						return nil
					})
				m.refreshTokenRepository.EXPECT().
					Save(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, hash string, payload refreshtoken.RefreshTokenPayload) error {
						assert.NotEmpty(t, hash)
						assert.Equal(t, testSubjectID, payload.SubjectID)
						assert.Equal(t, role.Specialist.String(), payload.Role)
						assert.Greater(t, payload.TTL, time.Duration(0))
						return nil
					})
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, evt event.Event) error {
						assert.Equal(t, authevents.AuthCredentialActivated, evt.Name)
						payload := evt.Payload.(map[string]any)
						assert.Equal(t, testSubjectID, payload["subject_id"])
						assert.Equal(t, role.Specialist.String(), payload["role"])
						assert.Equal(t, testEmail, payload["email"])
						return nil
					})
			},
			validateResult: func(t *testing.T, result *application.SetPasswordResult) {
				require.NotNil(t, result)
				require.NotNil(t, result.TokenPair)
				assert.Equal(t, testAccessJWT, result.TokenPair.AccessToken)
				assert.Equal(t, testRefreshOpaque, result.TokenPair.RefreshToken)
				assert.Equal(t, testSubjectID, result.SubjectID)
				assert.Equal(t, role.Specialist, result.Role)
			},
		},
		{
			name:  "failure - tokenValidator retorna erro: mapeia para ErrInvalidSetPasswordToken e nao toca em colaboradores",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().
					Validate(gomock.Any(), testRawToken).
					Times(1).
					Return(nil, errValidator)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), gomock.Any()).Times(0)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				m.credentialRepository.EXPECT().UpdateWithSessionInTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				m.accessTokenIssuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(0)
				m.refreshTokenRepository.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				m.eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: application.ErrInvalidSetPasswordToken,
		},
		{
			name:  "failure - Consume retorna erro: mapeia para ErrFailedToConsumeSingleUse",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(false, errRedisDown)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				expectLoggerAnyError(m.logger, 1)
			},
			expectError: true,
			expectedErr: application.ErrFailedToConsumeSingleUse,
		},
		{
			name:  "failure - Consume retorna false (ja consumido): mapeia para ErrSingleUseTokenAlreadyUsed",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(false, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: application.ErrSingleUseTokenAlreadyUsed,
		},
		{
			name:  "failure - FindBySubjectAndRole retorna erro: mapeia para ErrFailedToFindCredential",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(nil, errPostgresDown)
				m.credentialRepository.EXPECT().UpdateWithSessionInTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				expectLoggerAnyError(m.logger, 1)
			},
			expectError: true,
			expectedErr: application.ErrFailedToFindCredential,
		},
		{
			name:  "failure - FindBySubjectAndRole retorna nil: mapeia para ErrCredentialNotFound",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(nil, nil)
				m.credentialRepository.EXPECT().UpdateWithSessionInTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: application.ErrCredentialNotFound,
		},
		{
			name:  "failure - credential com status active: mapeia para ErrCredentialNotPending",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				cred := pendingCredentialFactory(func(c *credential.Credential) { c.Status = credential.StatusActive })
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(cred, nil)
				m.credentialRepository.EXPECT().UpdateWithSessionInTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: application.ErrCredentialNotPending,
		},
		{
			name:  "failure - credential com status locked: mapeia para ErrCredentialNotPending",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				cred := pendingCredentialFactory(func(c *credential.Credential) { c.Status = credential.StatusLocked })
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(cred, nil)
			},
			expectError: true,
			expectedErr: application.ErrCredentialNotPending,
		},
		{
			name:  "failure - password curta: retorna password.ErrTooShort sem hashear",
			input: inputFactory(func(in *application.SetPasswordDTO) { in.Password = testWeakShort }),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(pendingCredentialFactory(), nil)
				m.accessTokenIssuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: password.ErrTooShort,
		},
		{
			name:  "failure - password sem digitos: retorna password.ErrMissingRequiredChars",
			input: inputFactory(func(in *application.SetPasswordDTO) { in.Password = testWeakNoDigit }),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(pendingCredentialFactory(), nil)
				m.accessTokenIssuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: password.ErrMissingRequiredChars,
		},
		{
			name:  "failure - IssueAccessAndRefresh retorna erro: mapeia para ErrFailedToIssueTokenPair",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(pendingCredentialFactory(), nil)
				m.accessTokenIssuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(1).Return(nil, errIssuer)
				m.credentialRepository.EXPECT().UpdateWithSessionInTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				expectLoggerAnyError(m.logger, 1)
			},
			expectError: true,
			expectedErr: application.ErrFailedToIssueTokenPair,
		},
		{
			name:  "failure - UpdateWithSessionInTransaction retorna erro: mapeia para ErrFailedToPersistCredential",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(pendingCredentialFactory(), nil)
				m.accessTokenIssuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(1).Return(issuedTokenPairFactory(), nil)
				m.credentialRepository.EXPECT().UpdateWithSessionInTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errPostgresDown)
				m.refreshTokenRepository.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				expectLoggerAnyError(m.logger, 1)
			},
			expectError: true,
			expectedErr: application.ErrFailedToPersistCredential,
		},
		{
			name:  "failure - RefreshToken.Save retorna erro: mapeia para ErrFailedToCacheRefreshToken",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(pendingCredentialFactory(), nil)
				m.accessTokenIssuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(1).Return(issuedTokenPairFactory(), nil)
				m.credentialRepository.EXPECT().UpdateWithSessionInTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.refreshTokenRepository.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errRefreshCache)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
				m.eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)
				expectLoggerAnyError(m.logger, 1)
			},
			expectError: true,
			expectedErr: application.ErrFailedToCacheRefreshToken,
		},
		{
			name:  "happy path - audit retorna erro mas fluxo principal segue (fire-and-forget)",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(pendingCredentialFactory(), nil)
				m.accessTokenIssuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(1).Return(issuedTokenPairFactory(), nil)
				m.credentialRepository.EXPECT().UpdateWithSessionInTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.refreshTokenRepository.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(errAudit)
				m.eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				expectLoggerAnyError(m.logger, 1)
			},
			validateResult: func(t *testing.T, result *application.SetPasswordResult) {
				require.NotNil(t, result)
				assert.Equal(t, testAccessJWT, result.TokenPair.AccessToken)
			},
		},
		{
			name:  "happy path - event dispatch retorna erro mas fluxo principal segue (fire-and-forget)",
			input: inputFactory(),
			setupMocks: func(m *useCaseMocks) {
				m.tokenValidator.EXPECT().Validate(gomock.Any(), testRawToken).Times(1).Return(validatedTokenFactory(), nil)
				m.singleUseTokenRepository.EXPECT().Consume(gomock.Any(), testJTI).Times(1).Return(true, nil)
				m.credentialRepository.EXPECT().FindBySubjectAndRole(gomock.Any(), testSubjectID, role.Specialist).Times(1).Return(pendingCredentialFactory(), nil)
				m.accessTokenIssuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(1).Return(issuedTokenPairFactory(), nil)
				m.credentialRepository.EXPECT().UpdateWithSessionInTransaction(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.refreshTokenRepository.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.eventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(1).Return(errEvent)
				expectLoggerAnyError(m.logger, 1)
			},
			validateResult: func(t *testing.T, result *application.SetPasswordResult) {
				require.NotNil(t, result)
				assert.Equal(t, testAccessJWT, result.TokenPair.AccessToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mocksBundle := newUseCaseMocks(ctrl)
			tt.setupMocks(mocksBundle)

			uc := mocksBundle.build()

			result, err := uc.Execute(context.Background(), tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
				return
			}
			require.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}
