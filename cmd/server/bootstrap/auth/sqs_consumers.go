package auth

import (
	"context"
	"database/sql"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/bootstrap/infra"
	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/email"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	createspecialistsqs "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/create_specialist_credential/adapters/inbound/sqs"
	sendcredentialssqs "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/send_credentials_email/adapters/inbound/sqs"
	tokenissuer "github.com/lgustavopalmieri/healing-specialist/internal/platform/tokenissuer"
)

const (
	ConsumerRegisterCredential   = "specialist-register-credential"
	ConsumerSendCredentialsEmail = "auth-send-credentials-email"
)

type SQSConsumerDependencies struct {
	AuthDB         *sql.DB
	RedisClient    *redis.Client
	Signer         *tokenissuer.Signer
	EventPublisher event.EventDispatcher
	EmailSender    email.EmailSender
	SQS            *infra.SQSResources
	Config         *config.Config
}

func InitSQSConsumers(ctx context.Context, deps SQSConsumerDependencies) {
	log.Println("Starting Auth SQS consumers...")

	createspecialistsqs.NewCreateSpecialistCredentialSQSManager(ctx, createspecialistsqs.ManagerDependencies{
		AuthDB:               deps.AuthDB,
		RedisClient:          deps.RedisClient,
		Signer:               deps.Signer,
		EventPublisher:       deps.EventPublisher,
		SQSClient:            deps.SQS.Client,
		SpecialistCreatedURL: deps.SQS.QueueURLs[ConsumerRegisterCredential],
		SetPasswordTTL:       deps.Config.Auth.SetPasswordTTL,
	})

	sendcredentialssqs.NewSendCredentialsEmailSQSManager(ctx, sendcredentialssqs.ManagerDependencies{
		EmailSender:          deps.EmailSender,
		SQSClient:            deps.SQS.Client,
		CredentialPendingURL: deps.SQS.QueueURLs[ConsumerSendCredentialsEmail],
		SetPasswordURL:       deps.Config.Auth.SetPasswordBaseURL,
	})

	log.Println("Auth SQS consumers started")
}
