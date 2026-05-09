package http_test

import (
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

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	authhttp "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/http"
	sharedmocks "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/shared/mocks"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func validSpecialistClaims(subjectID string, overrides ...func(*claims.Claims)) *claims.Claims {
	now := time.Now()
	c := &claims.Claims{
		Subject:  subjectID,
		Role:     role.Specialist,
		Email:    "specialist@healing.com",
		Provider: provider.Password,
		TokenID:  "jti-abc",
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

type httpCase struct {
	name           string
	method         string
	path           string
	registerPath   string
	buildRoutes    func(*policy.RoutePolicy)
	header         string
	setupMocks     func(*sharedmocks.MockValidateTokenUseCase)
	expectedStatus int
	expectedCode   string
	validateCtx    func(t *testing.T, ctxSeen context.Context)
}

func runHTTPCase(t *testing.T, tc httpCase) {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockValidator := sharedmocks.NewMockValidateTokenUseCase(ctrl)
	if tc.setupMocks != nil {
		tc.setupMocks(mockValidator)
	}

	routes := policy.New()
	if tc.buildRoutes != nil {
		tc.buildRoutes(routes)
	}

	enforcer := policy.NewLocalEnforcer()

	router := gin.New()
	router.Use(authhttp.Middleware(mockValidator, enforcer, routes))

	var capturedCtx context.Context
	handler := func(c *gin.Context) {
		capturedCtx = c.Request.Context()
		c.Status(http.StatusOK)
	}

	registerPath := tc.registerPath
	if registerPath == "" {
		registerPath = tc.path
	}

	switch tc.method {
	case http.MethodGet:
		router.GET(registerPath, handler)
	case http.MethodPost:
		router.POST(registerPath, handler)
	case http.MethodPatch:
		router.PATCH(registerPath, handler)
	case http.MethodDelete:
		router.DELETE(registerPath, handler)
	default:
		router.Handle(tc.method, registerPath, handler)
	}

	req := httptest.NewRequest(tc.method, tc.path, nil)
	if tc.header != "" {
		req.Header.Set(authhttp.AuthorizationHeader, tc.header)
	}
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, tc.expectedStatus, w.Code, "status code mismatch. body=%s", w.Body.String())

	if tc.expectedCode != "" {
		var body authhttp.ErrorResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, tc.expectedCode, body.Code)
	}

	if tc.validateCtx != nil && tc.expectedStatus == http.StatusOK {
		require.NotNil(t, capturedCtx)
		tc.validateCtx(t, capturedCtx)
	}
}

func TestMiddleware_HTTP(t *testing.T) {
	tests := []httpCase{
		{
			name:   "happy path - rota publica registrada retorna 200 sem validar token",
			method: http.MethodGet,
			path:   "/health",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPPublic("GET", "/health")
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "happy path - rota publica nao exige header Authorization",
			method: http.MethodPost,
			path:   "/api/v1/specialists",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPPublic("POST", "/api/v1/specialists")
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "happy path - rota autenticada com token valido e role permitida retorna 200 e injeta claims",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPAuthenticated("POST", "/api/v1/auth/logout", role.Specialist, role.Patient)
			},
			header: "Bearer valid-token",
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "valid-token").
					Times(1).
					Return(validSpecialistClaims("subject-1"), nil)
			},
			expectedStatus: http.StatusOK,
			validateCtx: func(t *testing.T, ctxSeen context.Context) {
				c, ok := claims.FromContext(ctxSeen)
				require.True(t, ok)
				assert.Equal(t, "subject-1", c.Subject)
				assert.Equal(t, role.Specialist, c.Role)
			},
		},
		{
			name:         "happy path - rota owned com claims.Subject igual ao path param retorna 200",
			method:       http.MethodPatch,
			path:         "/api/v1/specialists/subject-1",
			registerPath: "/api/v1/specialists/:id",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPOwner("PATCH", "/api/v1/specialists/:id", "id", role.Specialist)
			},
			header: "Bearer valid-token",
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "valid-token").
					Times(1).
					Return(validSpecialistClaims("subject-1"), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "failure - rota nao registrada no RoutePolicy retorna 403 (fail closed)",
			method: http.MethodGet,
			path:   "/api/v1/unknown",
			buildRoutes: func(rp *policy.RoutePolicy) {
				// intencionalmente nao registra nada
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusForbidden,
			expectedCode:   "forbidden",
		},
		{
			name:   "failure - rota autenticada sem header Authorization retorna 401",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPAuthenticated("POST", "/api/v1/auth/logout", role.Specialist)
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "unauthenticated",
		},
		{
			name:   "failure - rota autenticada com header sem prefixo Bearer retorna 401",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPAuthenticated("POST", "/api/v1/auth/logout", role.Specialist)
			},
			header: "Basic abc",
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "invalid_token",
		},
		{
			name:   "failure - rota autenticada com header Bearer vazio retorna 401",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPAuthenticated("POST", "/api/v1/auth/logout", role.Specialist)
			},
			header: "Bearer ",
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "invalid_token",
		},
		{
			name:   "failure - rota autenticada com token invalido retorna 401",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPAuthenticated("POST", "/api/v1/auth/logout", role.Specialist)
			},
			header: "Bearer bad-token",
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "bad-token").
					Times(1).
					Return(nil, autherrors.ErrInvalidToken)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "invalid_token",
		},
		{
			name:   "failure - rota autenticada com token expirado retorna 401 com code token_expired",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPAuthenticated("POST", "/api/v1/auth/logout", role.Specialist)
			},
			header: "Bearer expired-token",
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "expired-token").
					Times(1).
					Return(nil, autherrors.ErrExpiredToken)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "token_expired",
		},
		{
			name:   "failure - rota autenticada com token blacklisted retorna 401 com code token_revoked",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPAuthenticated("POST", "/api/v1/auth/logout", role.Specialist)
			},
			header: "Bearer revoked-token",
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "revoked-token").
					Times(1).
					Return(nil, autherrors.ErrBlacklistedToken)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "token_revoked",
		},
		{
			name:   "failure - rota autenticada com role diferente do allowed retorna 403",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPAuthenticated("POST", "/api/v1/auth/logout", role.Admin)
			},
			header: "Bearer valid-token",
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "valid-token").
					Times(1).
					Return(validSpecialistClaims("subject-1"), nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedCode:   "forbidden_wrong_role",
		},
		{
			name:         "failure - rota owned com subject diferente do path param retorna 403",
			method:       http.MethodPatch,
			path:         "/api/v1/specialists/other-subject",
			registerPath: "/api/v1/specialists/:id",
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.HTTPOwner("PATCH", "/api/v1/specialists/:id", "id", role.Specialist)
			},
			header: "Bearer valid-token",
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "valid-token").
					Times(1).
					Return(validSpecialistClaims("subject-1"), nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedCode:   "forbidden_not_owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runHTTPCase(t, tt)
		})
	}
}
