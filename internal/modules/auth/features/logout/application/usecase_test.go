package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/logout/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/logout/application/mocks"
)

var errInfra = errors.New("infra failure")

func inputFactory() application.LogoutDTO {
	return application.LogoutDTO{
		RefreshToken:   "refresh-opaque",
		AccessTokenJTI: "access-jti-123",
		AccessTokenExp: time.Now().Add(50 * time.Minute),
		SubjectID:      "subject-1",
		Role:           "specialist",
		IPAddress:      "1.2.3.4",
		UserAgent:      "go-test",
	}
}

func TestLogoutUseCase_Execute(t *testing.T) {
	tests := []struct {
		name       string
		input      application.LogoutDTO
		setupMocks func(
			refreshRepo *mocks.MockRefreshTokenRepository,
			blacklistRepo *mocks.MockBlacklistRepository,
			sessRepo *mocks.MockSessionRepository,
			auditRepo *mocks.MockAuditRepository,
			logger *mocks.MockLogger,
		)
		expectError bool
		expectedErr error
	}{
		{
			name:  "happy path - deleta refresh, blacklista access, audit logout",
			input: inputFactory(),
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, blacklistRepo *mocks.MockBlacklistRepository, sessRepo *mocks.MockSessionRepository, auditRepo *mocks.MockAuditRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				blacklistRepo.EXPECT().Blacklist(gomock.Any(), "access-jti-123", gomock.Any()).Times(1).Return(nil)
				auditRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
		},
		{
			name:  "failure - Delete refresh falha retorna ErrDeleteRefreshToken",
			input: inputFactory(),
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, blacklistRepo *mocks.MockBlacklistRepository, sessRepo *mocks.MockSessionRepository, auditRepo *mocks.MockAuditRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(errInfra)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: application.ErrDeleteRefreshToken,
		},
		{
			name:  "failure - Blacklist access falha retorna ErrBlacklistAccessToken",
			input: inputFactory(),
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, blacklistRepo *mocks.MockBlacklistRepository, sessRepo *mocks.MockSessionRepository, auditRepo *mocks.MockAuditRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				blacklistRepo.EXPECT().Blacklist(gomock.Any(), "access-jti-123", gomock.Any()).Times(1).Return(errInfra)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
			expectError: true,
			expectedErr: application.ErrBlacklistAccessToken,
		},
		{
			name:  "happy path - audit falha NAO propaga",
			input: inputFactory(),
			setupMocks: func(refreshRepo *mocks.MockRefreshTokenRepository, blacklistRepo *mocks.MockBlacklistRepository, sessRepo *mocks.MockSessionRepository, auditRepo *mocks.MockAuditRepository, logger *mocks.MockLogger) {
				refreshRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				blacklistRepo.EXPECT().Blacklist(gomock.Any(), "access-jti-123", gomock.Any()).Times(1).Return(nil)
				auditRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(errInfra)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			refreshRepo := mocks.NewMockRefreshTokenRepository(ctrl)
			blacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
			sessRepo := mocks.NewMockSessionRepository(ctrl)
			auditRepo := mocks.NewMockAuditRepository(ctrl)
			logger := mocks.NewMockLogger(ctrl)

			tt.setupMocks(refreshRepo, blacklistRepo, sessRepo, auditRepo, logger)

			uc := application.NewLogoutUseCase(application.LogoutUseCaseDependencies{
				RefreshTokenRepository: refreshRepo,
				BlacklistRepository:    blacklistRepo,
				SessionRepository:      sessRepo,
				AuditRepository:        auditRepo,
				Logger:                 logger,
			})

			err := uc.Execute(context.Background(), tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
