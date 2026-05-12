package grpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	authgrpc "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/grpc"
	sharedmocks "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/shared/mocks"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func validClaims(subjectID string, r role.Role) *claims.Claims {
	now := time.Now()
	return &claims.Claims{
		Subject:  subjectID,
		Role:     r,
		Email:    "user@healing.com",
		Provider: provider.Password,
		TokenID:  "jti-xyz",
		IssuedAt: now.Add(-5 * time.Minute),
		ExpireAt: now.Add(55 * time.Minute),
		Issuer:   "healing-specialist",
		Audience: "healing-platform",
	}
}

func incomingCtxWithToken(token string) context.Context {
	md := metadata.Pairs(authgrpc.AuthorizationMetadataKey, "Bearer "+token)
	return metadata.NewIncomingContext(context.Background(), md)
}

type interceptorCase struct {
	name             string
	fullMethod       string
	ctx              context.Context
	buildRoutes      func(*policy.RoutePolicy)
	setupMocks       func(*sharedmocks.MockValidateTokenUseCase)
	expectedCode     codes.Code
	expectHandlerHit bool
	validateCtx      func(t *testing.T, handlerCtx context.Context)
}

func runInterceptorCase(t *testing.T, tc interceptorCase) {
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

	interceptor := authgrpc.UnaryInterceptor(mockValidator, enforcer, routes)

	var handlerHit bool
	var handlerCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerHit = true
		handlerCtx = ctx
		return "ok", nil
	}

	resp, err := interceptor(tc.ctx, struct{}{}, &grpc.UnaryServerInfo{FullMethod: tc.fullMethod}, handler)

	if tc.expectedCode == codes.OK {
		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
	} else {
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, tc.expectedCode, st.Code(), "expected %s, got %s: %s", tc.expectedCode, st.Code(), err.Error())
	}

	assert.Equal(t, tc.expectHandlerHit, handlerHit)

	if tc.validateCtx != nil && handlerHit {
		tc.validateCtx(t, handlerCtx)
	}
}

func TestUnaryInterceptor(t *testing.T) {
	tests := []interceptorCase{
		{
			name:       "happy path - metodo publico registrado chama handler sem validar token",
			fullMethod: "/pb.SpecialistService/CreateSpecialist",
			ctx:        context.Background(),
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.GRPCPublic("/pb.SpecialistService/CreateSpecialist")
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedCode:     codes.OK,
			expectHandlerHit: true,
		},
		{
			name:       "happy path - metodo autenticado com token valido chama handler e injeta claims",
			fullMethod: "/pb.SpecialistService/DoStuff",
			ctx:        incomingCtxWithToken("valid-token"),
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.GRPCAuthenticated("/pb.SpecialistService/DoStuff", role.Specialist)
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "valid-token").
					Times(1).
					Return(validClaims("subject-1", role.Specialist), nil)
			},
			expectedCode:     codes.OK,
			expectHandlerHit: true,
			validateCtx: func(t *testing.T, handlerCtx context.Context) {
				c, ok := claims.FromContext(handlerCtx)
				require.True(t, ok)
				assert.Equal(t, "subject-1", c.Subject)
				assert.Equal(t, role.Specialist, c.Role)
			},
		},
		{
			name:       "happy path - metodo owned so valida role no interceptor (ownership vai ao use case)",
			fullMethod: "/pb.UpdateSpecialistService/UpdateSpecialist",
			ctx:        incomingCtxWithToken("valid-token"),
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.GRPCOwnerInPayload("/pb.UpdateSpecialistService/UpdateSpecialist", role.Specialist)
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "valid-token").
					Times(1).
					Return(validClaims("subject-1", role.Specialist), nil)
			},
			expectedCode:     codes.OK,
			expectHandlerHit: true,
			validateCtx: func(t *testing.T, handlerCtx context.Context) {
				c, ok := claims.FromContext(handlerCtx)
				require.True(t, ok)
				assert.Equal(t, "subject-1", c.Subject)
			},
		},
		{
			name:       "failure - metodo nao registrado retorna PermissionDenied (fail closed)",
			fullMethod: "/pb.Unknown/Method",
			ctx:        context.Background(),
			buildRoutes: func(rp *policy.RoutePolicy) {
				// nada registrado
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedCode:     codes.PermissionDenied,
			expectHandlerHit: false,
		},
		{
			name:       "failure - metodo autenticado sem metadata retorna Unauthenticated",
			fullMethod: "/pb.SpecialistService/DoStuff",
			ctx:        context.Background(),
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.GRPCAuthenticated("/pb.SpecialistService/DoStuff", role.Specialist)
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedCode:     codes.Unauthenticated,
			expectHandlerHit: false,
		},
		{
			name:       "failure - metodo autenticado sem authorization na metadata retorna Unauthenticated",
			fullMethod: "/pb.SpecialistService/DoStuff",
			ctx:        metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-trace", "abc")),
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.GRPCAuthenticated("/pb.SpecialistService/DoStuff", role.Specialist)
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedCode:     codes.Unauthenticated,
			expectHandlerHit: false,
		},
		{
			name:       "failure - metodo autenticado com authorization sem prefixo Bearer retorna Unauthenticated",
			fullMethod: "/pb.SpecialistService/DoStuff",
			ctx:        metadata.NewIncomingContext(context.Background(), metadata.Pairs(authgrpc.AuthorizationMetadataKey, "Basic abc")),
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.GRPCAuthenticated("/pb.SpecialistService/DoStuff", role.Specialist)
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedCode:     codes.Unauthenticated,
			expectHandlerHit: false,
		},
		{
			name:       "failure - metodo autenticado com token invalido retorna Unauthenticated",
			fullMethod: "/pb.SpecialistService/DoStuff",
			ctx:        incomingCtxWithToken("bad-token"),
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.GRPCAuthenticated("/pb.SpecialistService/DoStuff", role.Specialist)
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "bad-token").
					Times(1).
					Return(nil, autherrors.ErrInvalidToken)
			},
			expectedCode:     codes.Unauthenticated,
			expectHandlerHit: false,
		},
		{
			name:       "failure - metodo autenticado com token expirado retorna Unauthenticated",
			fullMethod: "/pb.SpecialistService/DoStuff",
			ctx:        incomingCtxWithToken("expired-token"),
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.GRPCAuthenticated("/pb.SpecialistService/DoStuff", role.Specialist)
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "expired-token").
					Times(1).
					Return(nil, autherrors.ErrExpiredToken)
			},
			expectedCode:     codes.Unauthenticated,
			expectHandlerHit: false,
		},
		{
			name:       "failure - metodo autenticado com role errada retorna PermissionDenied",
			fullMethod: "/pb.SpecialistService/DoStuff",
			ctx:        incomingCtxWithToken("valid-token"),
			buildRoutes: func(rp *policy.RoutePolicy) {
				rp.GRPCAuthenticated("/pb.SpecialistService/DoStuff", role.Admin)
			},
			setupMocks: func(m *sharedmocks.MockValidateTokenUseCase) {
				m.EXPECT().
					Execute(gomock.Any(), "valid-token").
					Times(1).
					Return(validClaims("subject-1", role.Specialist), nil)
			},
			expectedCode:     codes.PermissionDenied,
			expectHandlerHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterceptorCase(t, tt)
		})
	}
}
