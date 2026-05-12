package httphandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	tokenpair "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/token_pair"
	httphandler "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/adapters/inbound/http_handler"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/adapters/inbound/http_handler/mocks"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/application"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func successResult() *application.LoginResult {
	now := time.Now()
	return &application.LoginResult{
		TokenPair: &tokenpair.TokenPair{
			AccessToken:      "access.jwt",
			AccessJTI:        "jti",
			AccessExpiresAt:  now.Add(1 * time.Hour),
			RefreshToken:     "refresh-opaque",
			RefreshExpiresAt: now.Add(168 * time.Hour),
		},
		SubjectID: "subject-1",
		Role:      role.Specialist,
	}
}

func buildRouter(h *httphandler.LoginHTTPHandler) *gin.Engine {
	engine := gin.New()
	api := engine.Group("/api/v1")
	h.RegisterRoutes(api)
	return engine
}

func TestLoginHTTPHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           any
		setupMocks     func(uc *mocks.MockLoginUseCaseInterface)
		expectedStatus int
		validateBody   func(t *testing.T, raw []byte)
		expectedError  string
	}{
		{
			name:   "happy path - POST /auth/specialist/login 200",
			method: http.MethodPost,
			path:   "/api/v1/auth/specialist/login",
			body:   httphandler.LoginRequest{Email: "user@healing.com", Password: "abc12345", DeviceInfo: "web"},
			setupMocks: func(uc *mocks.MockLoginUseCaseInterface) {
				uc.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, in application.LoginDTO) (*application.LoginResult, error) {
						assert.Equal(t, "user@healing.com", in.Email)
						assert.Equal(t, "abc12345", in.Password)
						assert.Equal(t, role.Specialist.String(), in.ExpectedRole)
						return successResult(), nil
					})
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, raw []byte) {
				var resp httphandler.LoginResponse
				require.NoError(t, json.Unmarshal(raw, &resp))
				assert.Equal(t, "access.jwt", resp.TokenPair.AccessToken)
				assert.Equal(t, "Bearer", resp.TokenPair.TokenType)
				assert.Equal(t, "specialist", resp.Role)
			},
		},
		{
			name:   "happy path - POST /auth/patient/login 200 com role=patient",
			method: http.MethodPost,
			path:   "/api/v1/auth/patient/login",
			body:   httphandler.LoginRequest{Email: "patient@healing.com", Password: "abc12345"},
			setupMocks: func(uc *mocks.MockLoginUseCaseInterface) {
				uc.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, in application.LoginDTO) (*application.LoginResult, error) {
						assert.Equal(t, role.Patient.String(), in.ExpectedRole)
						r := successResult()
						r.Role = role.Patient
						return r, nil
					})
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, raw []byte) {
				var resp httphandler.LoginResponse
				require.NoError(t, json.Unmarshal(raw, &resp))
				assert.Equal(t, "patient", resp.Role)
			},
		},
		{
			name:   "happy path - POST /auth/admin/login 200 com role=admin",
			method: http.MethodPost,
			path:   "/api/v1/auth/admin/login",
			body:   httphandler.LoginRequest{Email: "admin@healing.com", Password: "abc12345"},
			setupMocks: func(uc *mocks.MockLoginUseCaseInterface) {
				uc.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, in application.LoginDTO) (*application.LoginResult, error) {
						assert.Equal(t, role.Admin.String(), in.ExpectedRole)
						r := successResult()
						r.Role = role.Admin
						return r, nil
					})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "failure - body sem email retorna 400",
			method:         http.MethodPost,
			path:           "/api/v1/auth/specialist/login",
			body:           map[string]string{"password": "abc12345"},
			setupMocks:     func(uc *mocks.MockLoginUseCaseInterface) { uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0) },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "failure - body sem password retorna 400",
			method:         http.MethodPost,
			path:           "/api/v1/auth/specialist/login",
			body:           map[string]string{"email": "user@healing.com"},
			setupMocks:     func(uc *mocks.MockLoginUseCaseInterface) { uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0) },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "failure - ErrInvalidCredentials retorna 401 com 'invalid credentials'",
			method: http.MethodPost,
			path:   "/api/v1/auth/specialist/login",
			body:   httphandler.LoginRequest{Email: "user@healing.com", Password: "wrong1234"},
			setupMocks: func(uc *mocks.MockLoginUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, application.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid credentials",
		},
		{
			name:   "failure - ErrCredentialLocked retorna 401 com 'account locked'",
			method: http.MethodPost,
			path:   "/api/v1/auth/specialist/login",
			body:   httphandler.LoginRequest{Email: "user@healing.com", Password: "abc12345"},
			setupMocks: func(uc *mocks.MockLoginUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, application.ErrCredentialLocked)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "account locked",
		},
		{
			name:   "failure - ErrIssueTokens retorna 500 com 'internal error'",
			method: http.MethodPost,
			path:   "/api/v1/auth/specialist/login",
			body:   httphandler.LoginRequest{Email: "user@healing.com", Password: "abc12345"},
			setupMocks: func(uc *mocks.MockLoginUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, application.ErrIssueTokens)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal error",
		},
		{
			name:   "failure - ErrPersistSession retorna 500 com 'internal error'",
			method: http.MethodPost,
			path:   "/api/v1/auth/specialist/login",
			body:   httphandler.LoginRequest{Email: "user@healing.com", Password: "abc12345"},
			setupMocks: func(uc *mocks.MockLoginUseCaseInterface) {
				uc.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(1).Return(nil, application.ErrPersistSession)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockLoginUseCaseInterface(ctrl)
			tt.setupMocks(mockUC)

			handler := httphandler.NewLoginHTTPHandler(mockUC)
			router := buildRouter(handler)

			raw, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(raw))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var body httphandler.ErrorResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, tt.expectedError, body.Error)
			}
			if tt.validateBody != nil {
				tt.validateBody(t, rec.Body.Bytes())
			}
		})
	}
}
