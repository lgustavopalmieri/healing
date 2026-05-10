package bootstrap

import (
	"context"
	"fmt"
	"log"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	platformsqs "github.com/lgustavopalmieri/healing-specialist/internal/platform/sqs"
)

type SQSResources struct {
	Client    *awssqs.Client
	QueueURLs map[string]string
}

func InitSQS(ctx context.Context, cfg *config.Config) (*SQSResources, error) {
	log.Printf("Connecting to SQS (region=%s, prefix=%s)...", cfg.SQS.Region, cfg.SQS.QueuePrefix)

	client, err := platformsqs.NewClient(ctx, platformsqs.Config{
		Region:      cfg.SQS.Region,
		QueuePrefix: cfg.SQS.QueuePrefix,
		Endpoint:    cfg.SQS.Endpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SQS client: %w", err)
	}

	log.Println("Ensuring SQS consumer queues...")
	queueURLs, err := platformsqs.EnsureConsumerQueues(ctx, client, cfg.SQS.QueuePrefix, platformsqs.DefaultConsumerQueueDefinitions())
	if err != nil {
		return nil, fmt.Errorf("failed to ensure SQS consumer queues: %w", err)
	}

	log.Println("Running SQS health check...")
	if err := platformsqs.HealthCheck(ctx, client, queueURLs); err != nil {
		return nil, fmt.Errorf("SQS health check failed: %w", err)
	}

	log.Println("SQS initialized successfully")

	return &SQSResources{
		Client:    client,
		QueueURLs: queueURLs,
	}, nil
}
