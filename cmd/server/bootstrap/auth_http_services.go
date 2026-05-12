package bootstrap

import (
	"database/sql"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	loginhttp "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/login/adapters/inbound/http_handler"
	logouthttp "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/logout/adapters/inbound/http_handler"
	refreshhttp "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/refresh-token/adapters/inbound/http_handler"
	setpasswordhttp "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/adapters/inbound/http_handler"
	accesstokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/providers/access_token_issuer"
	auditrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/audit"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/blacklist"
	credentialrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/credential"
	refreshtokenrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/refresh_token"
	sessionrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/session"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/server"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

type AuthHTTPDependencies struct {
	AuthDB         *sql.DB
	RedisClient    *redis.Client
	Signer         *tokenissuer.Signer
	Keyring        *tokenissuer.Keyring
	EventPublisher event.EventDispatcher
	Logger         observability.Logger
	Config         *config.Config
}

func RegisterAuthHTTPServices(httpServer *server.HTTPServer, deps AuthHTTPDependencies) {
	log.Println("🔧 Registering Auth HTTP services...")

	api := httpServer.Engine.Group("/api/v1")

	setPasswordHandler := setpasswordhttp.NewSetPasswordHandler(setpasswordhttp.Dependencies{
		AuthDB:         deps.AuthDB,
		RedisClient:    deps.RedisClient,
		Signer:         deps.Signer,
		Keyring:        deps.Keyring,
		EventPublisher: deps.EventPublisher,
		Logger:         deps.Logger,
		Config:         deps.Config,
	})
	setPasswordHandler.RegisterRoutes(api)

	loginHandler := loginhttp.NewLoginHandler(loginhttp.Dependencies{
		AuthDB:      deps.AuthDB,
		RedisClient: deps.RedisClient,
		Signer:      deps.Signer,
		Logger:      deps.Logger,
		Config:      deps.Config,
	})
	loginHandler.RegisterRoutes(api)

	refreshTokenRepo := refreshtokenrepo.NewRefreshTokenCacheRepository(deps.RedisClient)
	credRepo := credentialrepo.NewCredentialDatabaseRepository(deps.AuthDB)
	sessRepo := sessionrepo.NewSessionDatabaseRepository(deps.AuthDB)
	auditRepo := auditrepo.NewAuditDatabaseRepository(deps.AuthDB)
	blacklistRepo := blacklist.NewBlacklistCacheRepository(deps.RedisClient)

	accessIssuer := accesstokenissuer.NewAccessTokenIssuer(accesstokenissuer.AccessTokenIssuerConfig{
		Signer:          deps.Signer,
		AccessTokenTTL:  deps.Config.Auth.AccessTokenTTL,
		RefreshTokenTTL: deps.Config.Auth.RefreshTokenTTL,
	})

	refreshHandler := refreshhttp.NewRefreshTokenHTTPHandlerFromDeps(refreshhttp.Dependencies{
		RefreshTokenRepository: refreshTokenRepo,
		AccessTokenIssuer:      accessIssuer,
		CredentialRepository:   credRepo,
		SessionRepository:      sessRepo,
		Logger:                 deps.Logger,
	})
	refreshHandler.RegisterRoutes(api)

	logoutHandler := logouthttp.NewLogoutHTTPHandlerFromDeps(logouthttp.Dependencies{
		RefreshTokenRepository: refreshTokenRepo,
		BlacklistRepository:    blacklistRepo,
		SessionRepository:      sessRepo,
		AuditRepository:        auditRepo,
		Logger:                 deps.Logger,
	})
	logoutHandler.RegisterRoutes(api)

	log.Println("✅ Auth HTTP services registered")
}
