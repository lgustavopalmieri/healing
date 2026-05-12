package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/password"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/application/mocks"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

const (
	testEmail     = "specialist@healing.com"
	testPassword  = "abc12345"
	testSubjectID = "subject-uuid-abc"
	testCredID    = "cred-uuid-xyz"
	testAccessJWT = "access.jwt.value"
	testRefresh   = "opaque-refresh-token"
)

var (
	errInfra = errors.New("infra failure")
)

type loginMocks struct {
	credentialRepository   *mocks.MockCredentialRepository
	accessTokenIssuer      *mocks.MockAccessTokenIssuer
	sessionRepository      *mocks.MockSessionRepository
	refreshTokenRepository *mocks.MockRefreshTokenRepository
	loginAttemptsTracker   *mocks.MockLoginAttemptsTracker
	auditRepository        *mocks.MockAuditRepository
	logger                 *mocks.MockLogger
}

func newLoginMocks(ctrl *gomock.Controller) *loginMocks {
	return &loginMocks{
		credentialRepository:   mocks.NewMockCredentialRepository(ctrl),
		accessTokenIssuer:      mocks.NewMockAccessTokenIssuer(ctrl),
		sessionRepository:      mocks.NewMockSessionRepository(ctrl),
		refreshTokenRepository: mocks.NewMockRefreshTokenRepository(ctrl),
		loginAttemptsTracker:   mocks.NewMockLoginAttemptsTracker(ctrl),
		auditRepository:        mocks.NewMockAuditRepository(ctrl),
		logger:                 mocks.NewMockLogger(ctrl),
	}
}

func (m *loginMocks) build() *application.LoginUseCase {
	return application.NewLoginUseCase(application.LoginUseCaseDependencies{
		CredentialRepository:   m.credentialRepository,
		AccessTokenIssuer:      m.accessTokenIssuer,
		SessionRepository:      m.sessionRepository,
		RefreshTokenRepository: m.refreshTokenRepository,
		LoginAttemptsTracker:   m.loginAttemptsTracker,
		AuditRepository:        m.auditRepository,
		Logger:                 m.logger,
	})
}

func activeCredentialFactory() *credential.Credential {
	hashed, _ := password.NewPassword(testPassword, password.ValidationConfig{MinLength: 1})
	h, _ := hashed.Hash(4)
	cred := credential.NewCredential(credential.NewCredentialInput{
		SubjectID: testSubjectID,
		Role:      role.Specialist,
		Provider:  provider.Password,
		Email:     testEmail,
	})
	cred.ID = testCredID
	_ = cred.Activate(h)
	return cred
}

func issuedTokenPairFactory() *tokenpair.TokenPair {
	now := time.Now()
	return &tokenpair.TokenPair{
		AccessToken:      testAccessJWT,
		AccessJTI:        "jti-access",
		AccessExpiresAt:  now.Add(1 * time.Hour),
		RefreshToken:     testRefresh,
		RefreshExpiresAt: now.Add(168 * time.Hour),
	}
}

func inputFactory(overrides ...func(*application.LoginDTO)) application.LoginDTO {
	in := application.LoginDTO{
		Email:        testEmail,
		Password:     testPassword,
		ExpectedRole: role.Specialist.String(),
		DeviceInfo:   "web",
		IPAddress:    "1.2.3.4",
		UserAgent:    "go-test",
	}
	for _, o := range overrides {
		o(&in)
	}
	return in
}

func TestLoginUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		input          application.LoginDTO
		setupMocks     func(m *loginMocks)
		expectError    bool
		expectedErr    error
		validateResult func(t *testing.T, result *application.LoginResult)
	}{
		{
			name:  "happy path - specialist login com credenciais validas",
			input: inputFactory(),
			setupMocks: func(m *loginMocks) {
				cred := activeCredentialFactory()
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(cred, nil)
				m.accessTokenIssuer.EXPECT().
					IssueAccessAndRefresh(gomock.Any(), gomock.Any()).
					Times(1).Return(issuedTokenPairFactory(), nil)
				m.sessionRepository.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, s *session.Session) error {
						assert.Equal(t, testSubjectID, s.SubjectID)
						assert.NotEmpty(t, s.RefreshTokenHash)
						return nil
					})
				m.refreshTokenRepository.EXPECT().
					Save(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, hash string, p refreshtoken.RefreshTokenPayload) error {
						assert.NotEmpty(t, hash)
						assert.Equal(t, testSubjectID, p.SubjectID)
						return nil
					})
				m.loginAttemptsTracker.EXPECT().Reset(gomock.Any(), testEmail).Times(1).Return(nil)
				m.credentialRepository.EXPECT().UpdateLastUsed(gomock.Any(), testCredID).Times(1).Return(nil)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
			validateResult: func(t *testing.T, result *application.LoginResult) {
				require.NotNil(t, result)
				assert.Equal(t, testAccessJWT, result.TokenPair.AccessToken)
				assert.Equal(t, testRefresh, result.TokenPair.RefreshToken)
				assert.Equal(t, testSubjectID, result.SubjectID)
				assert.Equal(t, role.Specialist, result.Role)
			},
		},
		{
			name:  "failure - role invalida retorna ErrInvalidCredentials sem tocar repos",
			input: inputFactory(func(in *application.LoginDTO) { in.ExpectedRole = "superadmin" }),
			setupMocks: func(m *loginMocks) {
				m.credentialRepository.EXPECT().FindByEmailProviderRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
			expectedErr: application.ErrInvalidCredentials,
		},
		{
			name:  "failure - email nao existe retorna ErrInvalidCredentials + audit not_found",
			input: inputFactory(),
			setupMocks: func(m *loginMocks) {
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(nil, nil)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
			expectError: true,
			expectedErr: application.ErrInvalidCredentials,
		},
		{
			name:  "failure - FindByEmailProviderRole retorna erro retorna ErrInvalidCredentials",
			input: inputFactory(),
			setupMocks: func(m *loginMocks) {
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(nil, errInfra)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: application.ErrInvalidCredentials,
		},
		{
			name:  "failure - credential locked retorna ErrCredentialLocked + audit locked",
			input: inputFactory(),
			setupMocks: func(m *loginMocks) {
				cred := activeCredentialFactory()
				cred.Status = credential.StatusLocked
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(cred, nil)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
			expectError: true,
			expectedErr: application.ErrCredentialLocked,
		},
		{
			name:  "failure - credential pending retorna ErrInvalidCredentials + audit status_pending",
			input: inputFactory(),
			setupMocks: func(m *loginMocks) {
				cred := credential.NewCredential(credential.NewCredentialInput{
					SubjectID: testSubjectID,
					Role:      role.Specialist,
					Provider:  provider.Password,
					Email:     testEmail,
				})
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(cred, nil)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
			expectError: true,
			expectedErr: application.ErrInvalidCredentials,
		},
		{
			name:  "failure - credential deleted retorna ErrInvalidCredentials + audit status_deleted",
			input: inputFactory(),
			setupMocks: func(m *loginMocks) {
				cred := activeCredentialFactory()
				cred.Status = credential.StatusDeleted
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(cred, nil)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
			expectError: true,
			expectedErr: application.ErrInvalidCredentials,
		},
		{
			name:  "failure - password errada retorna ErrInvalidCredentials + incrementa attempts + audit",
			input: inputFactory(func(in *application.LoginDTO) { in.Password = "wrongpassword1" }),
			setupMocks: func(m *loginMocks) {
				cred := activeCredentialFactory()
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(cred, nil)
				m.loginAttemptsTracker.EXPECT().Increment(gomock.Any(), testEmail).Times(1).Return(nil)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
			expectError: true,
			expectedErr: application.ErrInvalidCredentials,
		},
		{
			name:  "failure - IssueAccessAndRefresh retorna erro retorna ErrIssueTokens",
			input: inputFactory(),
			setupMocks: func(m *loginMocks) {
				cred := activeCredentialFactory()
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(cred, nil)
				m.accessTokenIssuer.EXPECT().
					IssueAccessAndRefresh(gomock.Any(), gomock.Any()).
					Times(1).Return(nil, errInfra)
				m.logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: application.ErrIssueTokens,
		},
		{
			name:  "failure - SessionRepository.Save retorna erro retorna ErrPersistSession",
			input: inputFactory(),
			setupMocks: func(m *loginMocks) {
				cred := activeCredentialFactory()
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(cred, nil)
				m.accessTokenIssuer.EXPECT().
					IssueAccessAndRefresh(gomock.Any(), gomock.Any()).
					Times(1).Return(issuedTokenPairFactory(), nil)
				m.sessionRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(errInfra)
				m.logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: application.ErrPersistSession,
		},
		{
			name:  "failure - RefreshTokenRepository.Save retorna erro retorna ErrCacheRefreshToken",
			input: inputFactory(),
			setupMocks: func(m *loginMocks) {
				cred := activeCredentialFactory()
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(cred, nil)
				m.accessTokenIssuer.EXPECT().
					IssueAccessAndRefresh(gomock.Any(), gomock.Any()).
					Times(1).Return(issuedTokenPairFactory(), nil)
				m.sessionRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.refreshTokenRepository.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errInfra)
				m.logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: application.ErrCacheRefreshToken,
		},
		{
			name:  "happy path - side effects fire-and-forget falham sem propagar",
			input: inputFactory(),
			setupMocks: func(m *loginMocks) {
				cred := activeCredentialFactory()
				m.credentialRepository.EXPECT().
					FindByEmailProviderRole(gomock.Any(), testEmail, provider.Password, role.Specialist).
					Times(1).Return(cred, nil)
				m.accessTokenIssuer.EXPECT().
					IssueAccessAndRefresh(gomock.Any(), gomock.Any()).
					Times(1).Return(issuedTokenPairFactory(), nil)
				m.sessionRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.refreshTokenRepository.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
				m.loginAttemptsTracker.EXPECT().Reset(gomock.Any(), testEmail).Times(1).Return(errInfra)
				m.credentialRepository.EXPECT().UpdateLastUsed(gomock.Any(), testCredID).Times(1).Return(errInfra)
				m.auditRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(errInfra)
			},
			validateResult: func(t *testing.T, result *application.LoginResult) {
				require.NotNil(t, result)
				assert.Equal(t, testAccessJWT, result.TokenPair.AccessToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := newLoginMocks(ctrl)
			tt.setupMocks(m)

			uc := m.build()
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
