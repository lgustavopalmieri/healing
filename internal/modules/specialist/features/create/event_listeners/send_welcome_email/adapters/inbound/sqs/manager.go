package sqs

import (
	"context"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/email"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/send_welcome_email/listener"
	platformsqs "github.com/lgustavopalmieri/healing-specialist/internal/platform/sqs"
)

type ManagerDependencies struct {
	EmailSender          email.EmailSender
	SQSClient            *awssqs.Client
	SpecialistCreatedURL string
}

func NewSendWelcomeEmailSQSManager(ctx context.Context, deps ManagerDependencies) {
	handler := listener.NewSendWelcomeEmailHandler(deps.EmailSender)

	consumer := platformsqs.NewSQSConsumer(
		deps.SQSClient,
		deps.SpecialistCreatedURL,
		listener.SpecialistCreatedEventName,
		handler,
	)

	go consumer.Start(ctx)
}
