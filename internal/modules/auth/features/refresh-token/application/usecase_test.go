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
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/refresh-token/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/refresh-token/application/mocks"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

var errInfra = errors.New("infra failure")

func issuedPair() *tokenpair.TokenPair {
	now := time.Now()
	return &tokenpair.TokenPair{
		AccessToken:      "new-access",
		AccessJTI:        "new-jti",
		AccessExpiresAt:  now.Add(1 * time.Hour),
		RefreshToken:     "new-refresh-opaque",
		RefreshExpiresAt: now.Add(168 * time.Hour),
	}
}

func storedPayload() *refreshtoken.RefreshTokenPayload {
	return &refreshtoken.RefreshTokenPayload{
		SessionID: "sess-1",
		SubjectID: "subject-1",
		Role:      role.Specialist.String(),
		TTL:       168 * time.Hour,
	}
}

func activeCred() *credential.Credential {
	cred := credential.NewCredential(credential.NewCredentialInput{
		SubjectID: "subject-1",
		Role:      role.Specialist,
		Provider:  provider.Password,
		Email:     "user@healing.com",
	})
	_ = cred.Activate(password.NewHashedPassword("hash"))
	return cred
}

func TestRefreshTokenUseCase_Execute(t *testing.T) {
	tests := []struct {
		name       string
		input      application.RefreshTokenDTO
		setupMocks func(
			refreshRepo *mocks.MockRefreshTokenRepository,
			issuer *mocks.MockAccessTokenIssuer,
			credRepo *mocks.MockCredentialRepository,
			sessRepo *mocks.MockSessionRepository,
			logger *mocks.MockLogger,
		)
		expectError bool
		expectedErr error
		validate    func(t *testing.T, result *application.RefreshTokenResult)
	}{
		{
			name:  "happy path - rotaciona refresh e retorna novo par",
			input: application.RefreshTokenDTO{RefreshToken: "old-refresh"},
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, issuer *mocks.MockAccessTokenIssuer, credRepo *mocks.MockCredentialRepository, sessRepo *mocks.MockSessionRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(storedPayload(), nil)
				refreshRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				credRepo.EXPECT().FindBySubjectAndRole(gomock.Any(), "subject-1", role.Specialist).Times(1).Return(activeCred(), nil)
				issuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(1).Return(issuedPair(), nil)
				refreshRepo.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
				sessRepo.EXPECT().UpdateRefreshTokenHash(gomock.Any(), "sess-1", gomock.Any()).Times(1).Return(nil)
			},
			validate: func(t *testing.T, result *application.RefreshTokenResult) {
				require.NotNil(t, result)
				assert.Equal(t, "new-access", result.TokenPair.AccessToken)
				assert.Equal(t, "new-refresh-opaque", result.TokenPair.RefreshToken)
				assert.Equal(t, "subject-1", result.SubjectID)
				assert.Equal(t, role.Specialist, result.Role)
			},
		},
		{
			name:  "failure - Find retorna nil retorna ErrInvalidRefreshToken",
			input: application.RefreshTokenDTO{RefreshToken: "unknown"},
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, issuer *mocks.MockAccessTokenIssuer, credRepo *mocks.MockCredentialRepository, sessRepo *mocks.MockSessionRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(nil, nil)
			},
			expectError: true,
			expectedErr: application.ErrInvalidRefreshToken,
		},
		{
			name:  "failure - Find retorna erro retorna ErrInvalidRefreshToken",
			input: application.RefreshTokenDTO{RefreshToken: "any"},
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, issuer *mocks.MockAccessTokenIssuer, credRepo *mocks.MockCredentialRepository, sessRepo *mocks.MockSessionRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(nil, errInfra)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: application.ErrInvalidRefreshToken,
		},
		{
			name:  "failure - Delete falha retorna ErrDeleteOldRefresh",
			input: application.RefreshTokenDTO{RefreshToken: "old"},
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, issuer *mocks.MockAccessTokenIssuer, credRepo *mocks.MockCredentialRepository, sessRepo *mocks.MockSessionRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(storedPayload(), nil)
				refreshRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(errInfra)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: application.ErrDeleteOldRefresh,
		},
		{
			name:  "failure - credential nao encontrada retorna ErrInvalidRefreshToken",
			input: application.RefreshTokenDTO{RefreshToken: "old"},
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, issuer *mocks.MockAccessTokenIssuer, credRepo *mocks.MockCredentialRepository, sessRepo *mocks.MockSessionRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(storedPayload(), nil)
				refreshRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				credRepo.EXPECT().FindBySubjectAndRole(gomock.Any(), "subject-1", role.Specialist).Times(1).Return(nil, nil)
			},
			expectError: true,
			expectedErr: application.ErrInvalidRefreshToken,
		},
		{
			name:  "failure - IssueAccessAndRefresh falha retorna ErrIssueNewTokens",
			input: application.RefreshTokenDTO{RefreshToken: "old"},
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, issuer *mocks.MockAccessTokenIssuer, credRepo *mocks.MockCredentialRepository, sessRepo *mocks.MockSessionRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(storedPayload(), nil)
				refreshRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				credRepo.EXPECT().FindBySubjectAndRole(gomock.Any(), "subject-1", role.Specialist).Times(1).Return(activeCred(), nil)
				issuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(1).Return(nil, errInfra)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: application.ErrIssueNewTokens,
		},
		{
			name:  "failure - Save novo refresh falha retorna ErrCacheNewRefresh",
			input: application.RefreshTokenDTO{RefreshToken: "old"},
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, issuer *mocks.MockAccessTokenIssuer, credRepo *mocks.MockCredentialRepository, sessRepo *mocks.MockSessionRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(storedPayload(), nil)
				refreshRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				credRepo.EXPECT().FindBySubjectAndRole(gomock.Any(), "subject-1", role.Specialist).Times(1).Return(activeCred(), nil)
				issuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(1).Return(issuedPair(), nil)
				refreshRepo.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errInfra)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: application.ErrCacheNewRefresh,
		},
		{
			name:  "happy path - UpdateRefreshTokenHash falha NAO propaga",
			input: application.RefreshTokenDTO{RefreshToken: "old"},
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, issuer *mocks.MockAccessTokenIssuer, credRepo *mocks.MockCredentialRepository, sessRepo *mocks.MockSessionRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(storedPayload(), nil)
				refreshRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				credRepo.EXPECT().FindBySubjectAndRole(gomock.Any(), "subject-1", role.Specialist).Times(1).Return(activeCred(), nil)
				issuer.EXPECT().IssueAccessAndRefresh(gomock.Any(), gomock.Any()).Times(1).Return(issuedPair(), nil)
				refreshRepo.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
				sessRepo.EXPECT().UpdateRefreshTokenHash(gomock.Any(), "sess-1", gomock.Any()).Times(1).Return(errInfra)
			},
			validate: func(t *testing.T, result *application.RefreshTokenResult) {
				require.NotNil(t, result)
				assert.Equal(t, "new-access", result.TokenPair.AccessToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			refreshRepo := mocks.NewMockRefreshTokenRepository(ctrl)
			issuer := mocks.NewMockAccessTokenIssuer(ctrl)
			credRepo := mocks.NewMockCredentialRepository(ctrl)
			sessRepo := mocks.NewMockSessionRepository(ctrl)
			logger := mocks.NewMockLogger(ctrl)

			tt.setupMocks(refreshRepo, issuer, credRepo, sessRepo, logger)

			uc := application.NewRefreshTokenUseCase(application.RefreshTokenUseCaseDependencies{
				RefreshTokenRepository: refreshRepo,
				AccessTokenIssuer:      issuer,
				CredentialRepository:   credRepo,
				SessionRepository:      sessRepo,
				Logger:                 logger,
			})

			result, err := uc.Execute(context.Background(), tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
				return
			}
			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
