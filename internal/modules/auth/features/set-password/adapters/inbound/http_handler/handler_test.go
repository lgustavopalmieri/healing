package httphandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/password"
	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	httphandler "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/adapters/inbound/http_handler"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/adapters/inbound/http_handler/mocks"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/application"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func validRequestBody() httphandler.SetPasswordRequest {
	return httphandler.SetPasswordRequest{
		Token:      "token-123",
		Password:   "abc12345",
		DeviceInfo: "web",
	}
}

func successResultFactory() *application.SetPasswordResult {
	now := time.Now()
	return &application.SetPasswordResult{
		TokenPair: &tokenpair.TokenPair{
			AccessToken:      "access.jwt",
			AccessJTI:        "jti-access",
			AccessExpiresAt:  now.Add(1 * time.Hour),
			RefreshToken:     "refresh-opaque",
			RefreshExpiresAt: now.Add(168 * time.Hour),
		},
		SubjectID: "subject-1",
		Role:      role.Specialist,
	}
}

func buildRouter(h *httphandler.SetPasswordHTTPHandler) *gin.Engine {
	engine := gin.New()
	api := engine.Group("/api/v1")
	h.RegisterRoutes(api)
	return engine
}

func TestSetPasswordHTTPHandler_SetPassword(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		contentType    string
		setupMocks     func(uc *mocks.MockSetPasswordUseCaseInterface)
		expectedStatus int
		expectedBody   string
		validateBody   func(t *testing.T, raw []byte)
	}{
		{
			name: "happy path - 200 com token pair e Bearer token_type",
			body: validRequestBody(),
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, in application.SetPasswordDTO) (*application.SetPasswordResult, error) {
						assert.Equal(t, "token-123", in.Token)
						assert.Equal(t, "abc12345", in.Password)
						assert.Equal(t, "web", in.DeviceInfo)
						assert.NotEmpty(t, in.IPAddress)
						return successResultFactory(), nil
					})
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, raw []byte) {
				var resp httphandler.SetPasswordResponse
				require.NoError(t, json.Unmarshal(raw, &resp))
				assert.Equal(t, "access.jwt", resp.TokenPair.AccessToken)
				assert.Equal(t, "refresh-opaque", resp.TokenPair.RefreshToken)
				assert.Equal(t, "Bearer", resp.TokenPair.TokenType)
				assert.Equal(t, "subject-1", resp.SubjectID)
				assert.Equal(t, "specialist", resp.Role)
			},
		},
		{
			name:        "failure - body nao-JSON retorna 400",
			body:        "not json at all",
			contentType: "text/plain",
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "failure - ErrInvalidSetPasswordToken mapeia para 401",
			body: validRequestBody(),
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, application.ErrInvalidSetPasswordToken)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   application.ErrInvalidSetPasswordToken.Error(),
		},
		{
			name: "failure - ErrSingleUseTokenAlreadyUsed mapeia para 401",
			body: validRequestBody(),
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, application.ErrSingleUseTokenAlreadyUsed)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   application.ErrSingleUseTokenAlreadyUsed.Error(),
		},
		{
			name: "failure - ErrCredentialNotFound mapeia para 404",
			body: validRequestBody(),
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, application.ErrCredentialNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   application.ErrCredentialNotFound.Error(),
		},
		{
			name: "failure - ErrCredentialNotPending mapeia para 409",
			body: validRequestBody(),
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, application.ErrCredentialNotPending)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   application.ErrCredentialNotPending.Error(),
		},
		{
			name: "failure - password.ErrTooShort mapeia para 400",
			body: validRequestBody(),
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, password.ErrTooShort)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   password.ErrTooShort.Error(),
		},
		{
			name: "failure - password.ErrMissingRequiredChars mapeia para 400",
			body: validRequestBody(),
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, password.ErrMissingRequiredChars)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   password.ErrMissingRequiredChars.Error(),
		},
		{
			name: "failure - ErrFailedToPersistCredential mapeia para 500",
			body: validRequestBody(),
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, application.ErrFailedToPersistCredential)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   application.ErrFailedToPersistCredential.Error(),
		},
		{
			name: "failure - ErrFailedToIssueTokenPair mapeia para 500",
			body: validRequestBody(),
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, application.ErrFailedToIssueTokenPair)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   application.ErrFailedToIssueTokenPair.Error(),
		},
		{
			name: "failure - erro nao-mapeado cai no default 500",
			body: validRequestBody(),
			setupMocks: func(uc *mocks.MockSetPasswordUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("unknown"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSetPasswordUseCaseInterface(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockUC)
			}

			handler := httphandler.NewSetPasswordHTTPHandler(mockUC)
			router := buildRouter(handler)

			var bodyReader *bytes.Buffer
			switch b := tt.body.(type) {
			case string:
				bodyReader = bytes.NewBufferString(b)
			default:
				raw, err := json.Marshal(b)
				require.NoError(t, err)
				bodyReader = bytes.NewBuffer(raw)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/set-password", bodyReader)
			contentType := tt.contentType
			if contentType == "" {
				contentType = "application/json"
			}
			req.Header.Set("Content-Type", contentType)
			req.Header.Set("User-Agent", "integration-test")
			req.RemoteAddr = "203.0.113.10:54321"

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedBody != "" {
				var body httphandler.ErrorResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, tt.expectedBody, body.Error)
			}
			if tt.validateBody != nil {
				tt.validateBody(t, rec.Body.Bytes())
			}
		})
	}
}
