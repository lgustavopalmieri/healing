package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	goredis "github.com/redis/go-redis/v9"

	redistest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/cache/redis"
	platformredis "github.com/lgustavopalmieri/healing-specialist/internal/platform/redis"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
	authhttp "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/http"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
	sdktoken "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"

	validatetokencache "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/validate-token/adapters/outbound/cache"
	validatetokenapp "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/validate-token/application"
)

const (
	testKID      = "test-kid"
	testIssuer   = "healing-specialist"
	testAudience = "healing-platform"
)

type testFixture struct {
	engine      *gin.Engine
	signer      *tokenissuer.Signer
	redisClient *goredis.Client
	cleanup     func()
}

func setupFixture(t *testing.T) *testFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyring := tokenissuer.NewKeyring(testKID, key)
	signer := tokenissuer.NewSigner(keyring, tokenissuer.SignerConfig{
		Issuer:   testIssuer,
		Audience: testAudience,
	})

	container := redistest.SetupRedisContainer(t)
	redisClient, err := platformredis.NewClient(context.Background(), container.Config())
	require.NoError(t, err)

	jwtValidator := sdktoken.NewJWTValidator(sdktoken.JWTValidatorConfig{
		PublicKeys: keyring.PublicKeys,
		Issuer:     testIssuer,
		Audience:   testAudience,
	})
	blacklistRepo := validatetokencache.NewBlacklistCacheRepository(redisClient)
	validateTokenUC := validatetokenapp.NewValidateTokenUseCase(jwtValidator, blacklistRepo)

	enforcer := policy.NewLocalEnforcer()

	rp := policy.New()
	rp.HTTPPublic("GET", "/health")
	rp.HTTPPublic("POST", "/api/v1/specialists")
	rp.HTTPAuthenticated("POST", "/api/v1/auth/logout", role.Specialist, role.Patient, role.Admin)
	rp.HTTPOwner("PATCH", "/api/v1/specialists/:id", "id", role.Specialist)

	engine := gin.New()
	engine.Use(authhttp.Middleware(validateTokenUC, enforcer, rp))

	engine.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.POST("/api/v1/specialists", func(c *gin.Context) { c.Status(http.StatusCreated) })
	engine.POST("/api/v1/auth/logout", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	engine.PATCH("/api/v1/specialists/:id", func(c *gin.Context) { c.Status(http.StatusOK) })

	return &testFixture{
		engine:      engine,
		signer:      signer,
		redisClient: redisClient,
		cleanup: func() {
			_ = redisClient.Close()
			container.Terminate(t)
		},
	}
}

func signAccessToken(t *testing.T, signer *tokenissuer.Signer, subjectID string, r role.Role) string {
	t.Helper()
	token, _, _, err := signer.SignAccess(tokenissuer.SignAccessInput{
		Subject:  subjectID,
		Role:     r,
		Email:    "user@healing.com",
		Provider: provider.Password,
		TTL:      1 * time.Hour,
	})
	require.NoError(t, err)
	return token
}

func TestAuthMiddleware_Integration(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		token          func(t *testing.T, f *testFixture) string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "rota publica GET /health — 200 sem token",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "rota publica POST /api/v1/specialists — 201 sem token",
			method:         http.MethodPost,
			path:           "/api/v1/specialists",
			body:           "{}",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "rota autenticada sem token — 401",
			method:         http.MethodPost,
			path:           "/api/v1/auth/logout",
			body:           `{"refresh_token":"x"}`,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "rota autenticada com token valido — 204",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			body:   `{"refresh_token":"x"}`,
			token: func(t *testing.T, f *testFixture) string {
				return signAccessToken(t, f.signer, "subject-1", role.Specialist)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "rota autenticada com token expirado — 401",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			body:   `{"refresh_token":"x"}`,
			token: func(t *testing.T, f *testFixture) string {
				tok, _, _, err := f.signer.SignAccess(tokenissuer.SignAccessInput{
					Subject:  "subject-1",
					Role:     role.Specialist,
					Email:    "user@healing.com",
					Provider: provider.Password,
					TTL:      -1 * time.Hour,
				})
				require.NoError(t, err)
				return tok
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "rota autenticada com token blacklisted — 401",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			body:   `{"refresh_token":"x"}`,
			token: func(t *testing.T, f *testFixture) string {
				tok, jti, _, err := f.signer.SignAccess(tokenissuer.SignAccessInput{
					Subject:  "subject-1",
					Role:     role.Specialist,
					Email:    "user@healing.com",
					Provider: provider.Password,
					TTL:      1 * time.Hour,
				})
				require.NoError(t, err)
				require.NoError(t, f.redisClient.Set(context.Background(), "auth:blacklist:"+jti, "1", 10*time.Second).Err())
				return tok
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "rota autenticada com role errada — 403",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			body:   `{"refresh_token":"x"}`,
			token: func(t *testing.T, f *testFixture) string {
				return signAccessToken(t, f.signer, "subject-1", role.Anonymous)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "rota owned — dono acessa — 200",
			method: http.MethodPatch,
			path:   "/api/v1/specialists/subject-1",
			body:   "{}",
			token: func(t *testing.T, f *testFixture) string {
				return signAccessToken(t, f.signer, "subject-1", role.Specialist)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "rota owned — outro subject — 403 forbidden_not_owner",
			method: http.MethodPatch,
			path:   "/api/v1/specialists/other-subject",
			body:   "{}",
			token: func(t *testing.T, f *testFixture) string {
				return signAccessToken(t, f.signer, "subject-1", role.Specialist)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "forbidden_not_owner",
		},
		{
			name:   "rota owned — role errada (patient tentando editar specialist) — 403",
			method: http.MethodPatch,
			path:   "/api/v1/specialists/subject-1",
			body:   "{}",
			token: func(t *testing.T, f *testFixture) string {
				return signAccessToken(t, f.signer, "subject-1", role.Patient)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "rota nao registrada — 403 fail-closed",
			method:         http.MethodDelete,
			path:           "/api/v1/unknown",
			expectedStatus: http.StatusForbidden,
		},
	}

	fixture := setupFixture(t)
	defer fixture.cleanup()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *strings.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			} else {
				bodyReader = strings.NewReader("")
			}

			req := httptest.NewRequest(tt.method, tt.path, bodyReader)
			req.Header.Set("Content-Type", "application/json")

			if tt.token != nil {
				tok := tt.token(t, fixture)
				req.Header.Set("Authorization", "Bearer "+tok)
			}

			rec := httptest.NewRecorder()
			fixture.engine.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code, "body: %s", rec.Body.String())

			if tt.expectedBody != "" {
				var body authhttp.ErrorResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Contains(t, body.Code, tt.expectedBody)
			}
		})
	}
}
