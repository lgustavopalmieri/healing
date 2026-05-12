package auth

import (
	"log"

	"github.com/redis/go-redis/v9"

	validatetokencache "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/validate-token/adapters/outbound/cache"
	validatetokenapp "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/validate-token/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/middleware/shared"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/policy"
	sdktoken "github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
)

type MiddlewareResources struct {
	ValidateTokenUseCase shared.ValidateTokenUseCase
	Enforcer             policy.Enforcer
	RoutePolicy          *policy.RoutePolicy
}

func InitMiddleware(keyring *tokenissuer.Keyring, redisClient *redis.Client, issuer, audience string) *MiddlewareResources {
	log.Println("Initializing auth middleware...")

	jwtValidator := sdktoken.NewJWTValidator(sdktoken.JWTValidatorConfig{
		PublicKeys: keyring.PublicKeys,
		Issuer:     issuer,
		Audience:   audience,
	})

	blacklistRepo := validatetokencache.NewBlacklistCacheRepository(redisClient)

	validateTokenUC := validatetokenapp.NewValidateTokenUseCase(jwtValidator, blacklistRepo)

	enforcer := policy.NewLocalEnforcer()
	routePolicy := BuildRoutePolicy()

	log.Println("Auth middleware initialized")

	return &MiddlewareResources{
		ValidateTokenUseCase: validateTokenUC,
		Enforcer:             enforcer,
		RoutePolicy:          routePolicy,
	}
}
