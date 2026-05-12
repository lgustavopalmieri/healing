package http

import (
	"github.com/gin-gonic/gin"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/claims"
	autherrors "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/errors"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/shared"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
)

func Middleware(
	validator shared.ValidateTokenUseCase,
	enforcer policy.Enforcer,
	routes *policy.RoutePolicy,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		rule, ok := routes.LookupHTTP(method, path)
		if !ok {
			WriteError(c, autherrors.ErrForbidden)
			return
		}

		if rule.Policy.AllowPublic {
			c.Next()
			return
		}

		rawToken, err := ExtractBearerToken(c.GetHeader(AuthorizationHeader))
		if err != nil {
			WriteError(c, err)
			return
		}

		validated, err := validator.Execute(c.Request.Context(), rawToken)
		if err != nil {
			WriteError(c, err)
			return
		}

		ctx := claims.WithClaims(c.Request.Context(), validated)
		c.Request = c.Request.WithContext(ctx)

		var ownerID string
		if rule.OwnerIDParam != "" {
			ownerID = c.Param(rule.OwnerIDParam)
		}

		if err := enforcer.Enforce(ctx, rule.Policy, validated, ownerID); err != nil {
			WriteError(c, err)
			return
		}

		c.Next()
	}
}
