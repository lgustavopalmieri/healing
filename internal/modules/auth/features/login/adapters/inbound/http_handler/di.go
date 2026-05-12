package httphandler

import (
	"database/sql"

	"github.com/redis/go-redis/v9"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/adapters/outbound/cache"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/application"
	accesstokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/providers/access_token_issuer"
	auditrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/audit"
	credentialrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/credential"
	refreshtoken "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	sessionrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/session"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

type Dependencies struct {
	AuthDB      *sql.DB
	RedisClient *redis.Client
	Signer      *tokenissuer.Signer
	Logger      observability.Logger
	Config      *config.Config
}

func NewLoginHandler(deps Dependencies) *LoginHTTPHandler {
	credentialRepository := credentialrepo.NewCredentialDatabaseRepository(deps.AuthDB)
	sessionRepository := sessionrepo.NewSessionDatabaseRepository(deps.AuthDB)
	auditRepository := auditrepo.NewAuditDatabaseRepository(deps.AuthDB)
	refreshTokenRepository := refreshtoken.NewRefreshTokenCacheRepository(deps.RedisClient)
	attemptsTracker := cache.NewLoginAttemptsTracker(deps.RedisClient)

	accessIssuer := accesstokenissuer.NewAccessTokenIssuer(accesstokenissuer.AccessTokenIssuerConfig{
		Signer:          deps.Signer,
		AccessTokenTTL:  deps.Config.Auth.AccessTokenTTL,
		RefreshTokenTTL: deps.Config.Auth.RefreshTokenTTL,
	})

	useCase := application.NewLoginUseCase(application.LoginUseCaseDependencies{
		CredentialRepository:   credentialRepository,
		AccessTokenIssuer:      accessIssuer,
		SessionRepository:      sessionRepository,
		RefreshTokenRepository: refreshTokenRepository,
		LoginAttemptsTracker:   attemptsTracker,
		AuditRepository:        auditRepository,
		Logger:                 deps.Logger,
	})

	return NewLoginHTTPHandler(useCase)
}
