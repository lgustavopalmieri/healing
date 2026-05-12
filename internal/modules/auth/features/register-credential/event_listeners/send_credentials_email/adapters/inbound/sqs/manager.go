package sqs

import (
	"context"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/email"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/send_credentials_email/listener"
	platformsqs "github.com/lgustavopalmieri/healing-specialist/internal/platform/sqs"
)

type ManagerDependencies struct {
	EmailSender          email.EmailSender
	SQSClient            *awssqs.Client
	CredentialPendingURL string
	SetPasswordURL       string
}

func NewSendCredentialsEmailSQSManager(ctx context.Context, deps ManagerDependencies) {
	handler := listener.NewSendCredentialsEmailHandler(deps.EmailSender, deps.SetPasswordURL)

	consumer := platformsqs.NewSQSConsumer(
		deps.SQSClient,
		deps.CredentialPendingURL,
		listener.AuthCredentialPendingEventName,
		handler,
	)

	go consumer.Start(ctx)
}
