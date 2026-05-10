package bootstrap

import (
	"context"
	"database/sql"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	createspecialistsqs "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/create_specialist_credential/adapters/inbound/sqs"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

type AuthSQSConsumerDependencies struct {
	AuthDB         *sql.DB
	RedisClient    *redis.Client
	Signer         *tokenissuer.Signer
	EventPublisher event.EventDispatcher
	SQS            *SQSResources
	Config         *config.Config
}

func InitAuthSQSConsumers(ctx context.Context, deps AuthSQSConsumerDependencies) {
	log.Println("Starting Auth SQS consumers...")

	createspecialistsqs.NewCreateSpecialistCredentialSQSManager(ctx, createspecialistsqs.ManagerDependencies{
		AuthDB:               deps.AuthDB,
		RedisClient:          deps.RedisClient,
		Signer:               deps.Signer,
		EventPublisher:       deps.EventPublisher,
		SQSClient:            deps.SQS.Client,
		SpecialistCreatedURL: deps.SQS.QueueURLs["specialist.created"],
		SetPasswordTTL:       deps.Config.Auth.SetPasswordTTL,
	})

	log.Println("Auth SQS consumers started")
}
