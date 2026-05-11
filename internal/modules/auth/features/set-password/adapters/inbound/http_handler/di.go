package httphandler

import (
	"database/sql"

	"github.com/redis/go-redis/v9"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/adapters/outbound/provider"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/adapters/outbound/singleuse"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/application"
	accesstokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/providers/access_token_issuer"
	singleusetoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/providers/single_use_token"
	auditrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/audit"
	credentialrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/credential"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

type Dependencies struct {
	AuthDB         *sql.DB
	RedisClient    *redis.Client
	Signer         *tokenissuer.Signer
	Keyring        *tokenissuer.Keyring
	EventPublisher event.EventDispatcher
	Logger         observability.Logger
	Config         *config.Config
}

func NewSetPasswordHandler(deps Dependencies) *SetPasswordHTTPHandler {
	credentialRepository := credentialrepo.NewCredentialDatabaseRepository(deps.AuthDB)
	auditRepository := auditrepo.NewAuditDatabaseRepository(deps.AuthDB)
	refreshTokenRepository := refreshtoken.NewRefreshTokenCacheRepository(deps.RedisClient)

	singleUseStore := singleusetoken.NewSingleUseTokenCacheRepository(deps.RedisClient)
	singleUseRepository := singleuse.NewSetPasswordSingleUseTokenRepository(singleUseStore)

	tokenValidator := provider.NewSetPasswordTokenValidator(provider.SetPasswordTokenValidatorConfig{
		Keyring:  deps.Keyring,
		Issuer:   deps.Config.Auth.Issuer,
		Audience: deps.Config.Auth.Audience,
	})

	accessIssuer := accesstokenissuer.NewAccessTokenIssuer(accesstokenissuer.AccessTokenIssuerConfig{
		Signer:          deps.Signer,
		AccessTokenTTL:  deps.Config.Auth.AccessTokenTTL,
		RefreshTokenTTL: deps.Config.Auth.RefreshTokenTTL,
	})

	useCase := application.NewSetPasswordUseCase(application.SetPasswordUseCaseDependencies{
		TokenValidator:           tokenValidator,
		SingleUseTokenRepository: singleUseRepository,
		CredentialRepository:     credentialRepository,
		AccessTokenIssuer:        accessIssuer,
		RefreshTokenRepository:   refreshTokenRepository,
		AuditRepository:          auditRepository,
		EventPublisher:           deps.EventPublisher,
		Logger:                   deps.Logger,
		PasswordMinLength:        deps.Config.Auth.PasswordMinLength,
		BcryptCost:               deps.Config.Auth.BcryptCost,
	})

	return NewSetPasswordHTTPHandler(useCase)
}
