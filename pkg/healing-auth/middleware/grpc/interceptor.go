package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/shared"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
)

func UnaryInterceptor(
	validator shared.ValidateTokenUseCase,
	enforcer policy.Enforcer,
	routes *policy.RoutePolicy,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		rule, ok := routes.LookupGRPC(info.FullMethod)
		if !ok {
			return nil, MapError(autherrors.ErrForbidden)
		}

		if rule.Policy.AllowPublic {
			return handler(ctx, req)
		}

		rawToken, err := ExtractBearerToken(ctx)
		if err != nil {
			return nil, MapError(err)
		}

		validated, err := validator.Execute(ctx, rawToken)
		if err != nil {
			return nil, MapError(err)
		}

		ctx = claims.WithClaims(ctx, validated)

		if err := enforcer.EnforceRoleOnly(ctx, rule.Policy, validated); err != nil {
			return nil, MapError(err)
		}

		return handler(ctx, req)
	}
}
