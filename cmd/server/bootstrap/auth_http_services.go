package bootstrap

import (
	"database/sql"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	setpasswordhttp "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/set-password/adapters/inbound/http_handler"
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

	log.Println("✅ Auth HTTP services registered")
}
