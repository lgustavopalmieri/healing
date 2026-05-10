package sqs

import (
	"context"
	"database/sql"
	"time"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/redis/go-redis/v9"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/create_specialist_credential/adapters/outbound/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/create_specialist_credential/adapters/outbound/tokens"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/create_specialist_credential/listener"
	platformsqs "github.com/lgustavopalmieri/healing-specialist/internal/platform/sqs"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

type ManagerDependencies struct {
	AuthDB               *sql.DB
	RedisClient          *redis.Client
	Signer               *tokenissuer.Signer
	EventPublisher       event.EventDispatcher
	SQSClient            *awssqs.Client
	SpecialistCreatedURL string
	SetPasswordTTL       time.Duration
}

func NewCreateSpecialistCredentialSQSManager(ctx context.Context, deps ManagerDependencies) {
	credentialRepository := database.NewCredentialDatabaseRepository(deps.AuthDB)
	setPasswordGenerator := tokens.NewSetPasswordTokenGenerator(deps.Signer, deps.RedisClient, deps.SetPasswordTTL)

	handler := listener.NewCreateSpecialistCredentialHandler(
		credentialRepository,
		setPasswordGenerator,
		deps.EventPublisher,
	)

	consumer := platformsqs.NewSQSConsumer(
		deps.SQSClient,
		deps.SpecialistCreatedURL,
		listener.SpecialistCreatedEventName,
		handler,
	)

	go consumer.Start(ctx)
}
