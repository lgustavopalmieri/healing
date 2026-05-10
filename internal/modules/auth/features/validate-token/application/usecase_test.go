package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/validate-token/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/validate-token/application/mocks"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

const (
	testRawToken = "raw.jwt.token"
	testJTI      = "jti-abc"
	testSubject  = "subject-1"
)

var errRedisDown = errors.New("redis connection refused")

func validClaimsFactory(overrides ...func(*claims.Claims)) *claims.Claims {
	now := time.Now()
	c := &claims.Claims{
		Subject:  testSubject,
		Role:     role.Specialist,
		Email:    "user@healing.com",
		Provider: provider.Password,
		TokenID:  testJTI,
		IssuedAt: now.Add(-5 * time.Minute),
		ExpireAt: now.Add(55 * time.Minute),
		Issuer:   "healing-specialist",
		Audience: "healing-platform",
	}
	for _, o := range overrides {
		o(c)
	}
	return c
}

type validateTokenCase struct {
	name           string
	rawToken       string
	setupMocks     func(v *mocks.MockTokenValidator, b *mocks.MockBlacklistRepository)
	expectError    bool
	expectedErr    error
	validateResult func(t *testing.T, got *claims.Claims)
}

func TestValidateTokenUseCase_Execute(t *testing.T) {
	tests := []validateTokenCase{
		{
			name:     "happy path - token valido e nao blacklisted retorna claims",
			rawToken: testRawToken,
			setupMocks: func(tokenValidator *mocks.MockTokenValidator, blacklistRepository *mocks.MockBlacklistRepository) {
				tokenValidator.EXPECT().
					Validate(gomock.Any(), testRawToken).
					Times(1).
					Return(validClaimsFactory(), nil)
				blacklistRepository.EXPECT().
					IsBlacklisted(gomock.Any(), testJTI).
					Times(1).
					Return(false, nil)
			},
			validateResult: func(t *testing.T, got *claims.Claims) {
				require.NotNil(t, got)
				assert.Equal(t, testSubject, got.Subject)
				assert.Equal(t, role.Specialist, got.Role)
				assert.Equal(t, testJTI, got.TokenID)
			},
		},
		{
			name:     "happy path - token valido com jti vazio retorna claims sem consultar blacklist",
			rawToken: testRawToken,
			setupMocks: func(tokenValidator *mocks.MockTokenValidator, blacklistRepository *mocks.MockBlacklistRepository) {
				tokenValidator.EXPECT().
					Validate(gomock.Any(), testRawToken).
					Times(1).
					Return(validClaimsFactory(func(c *claims.Claims) { c.TokenID = "" }), nil)
				blacklistRepository.EXPECT().
					IsBlacklisted(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validateResult: func(t *testing.T, got *claims.Claims) {
				require.NotNil(t, got)
				assert.Empty(t, got.TokenID)
			},
		},
		{
			name:     "failure - validator retorna ErrInvalidToken propaga o erro",
			rawToken: "bad.token",
			setupMocks: func(tokenValidator *mocks.MockTokenValidator, blacklistRepository *mocks.MockBlacklistRepository) {
				tokenValidator.EXPECT().
					Validate(gomock.Any(), "bad.token").
					Times(1).
					Return(nil, autherrors.ErrInvalidToken)
				blacklistRepository.EXPECT().
					IsBlacklisted(gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectError: true,
			expectedErr: autherrors.ErrInvalidToken,
		},
		{
			name:     "failure - validator retorna ErrExpiredToken propaga o erro",
			rawToken: "expired.token",
			setupMocks: func(tokenValidator *mocks.MockTokenValidator, blacklistRepository *mocks.MockBlacklistRepository) {
				tokenValidator.EXPECT().
					Validate(gomock.Any(), "expired.token").
					Times(1).
					Return(nil, autherrors.ErrExpiredToken)
				blacklistRepository.EXPECT().
					IsBlacklisted(gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectError: true,
			expectedErr: autherrors.ErrExpiredToken,
		},
		{
			name:     "failure - blacklist retorna erro e o usecase propaga",
			rawToken: testRawToken,
			setupMocks: func(tokenValidator *mocks.MockTokenValidator, blacklistRepository *mocks.MockBlacklistRepository) {
				tokenValidator.EXPECT().
					Validate(gomock.Any(), testRawToken).
					Times(1).
					Return(validClaimsFactory(), nil)
				blacklistRepository.EXPECT().
					IsBlacklisted(gomock.Any(), testJTI).
					Times(1).
					Return(false, errRedisDown)
			},
			expectError: true,
			expectedErr: errRedisDown,
		},
		{
			name:     "failure - token blacklisted retorna ErrBlacklistedToken",
			rawToken: testRawToken,
			setupMocks: func(tokenValidator *mocks.MockTokenValidator, blacklistRepository *mocks.MockBlacklistRepository) {
				tokenValidator.EXPECT().
					Validate(gomock.Any(), testRawToken).
					Times(1).
					Return(validClaimsFactory(), nil)
				blacklistRepository.EXPECT().
					IsBlacklisted(gomock.Any(), testJTI).
					Times(1).
					Return(true, nil)
			},
			expectError: true,
			expectedErr: autherrors.ErrBlacklistedToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tokenValidator := mocks.NewMockTokenValidator(ctrl)
			blacklistRepository := mocks.NewMockBlacklistRepository(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(tokenValidator, blacklistRepository)
			}

			uc := application.NewValidateTokenUseCase(tokenValidator, blacklistRepository)

			got, err := uc.Execute(context.Background(), tt.rawToken)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, got)
			}
		})
	}
}
