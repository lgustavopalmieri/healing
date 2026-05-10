package bootstrap

import (
	"context"
	"fmt"
	"log"

	awssns "github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	platformsns "github.com/lgustavopalmieri/healing-specialist/internal/platform/sns"
	platformsqs "github.com/lgustavopalmieri/healing-specialist/internal/platform/sqs"
)

type SNSResources struct {
	Client    *awssns.Client
	Producer  *platformsns.SNSProducer
	TopicARNs map[string]string
}

func InitSNS(ctx context.Context, cfg *config.Config, sqsResources *SQSResources) (*SNSResources, error) {
	log.Printf("Connecting to SNS (region=%s, prefix=%s)...", cfg.SNS.Region, cfg.SNS.TopicPrefix)

	client, err := platformsns.NewClient(ctx, platformsns.Config{
		Region:      cfg.SNS.Region,
		Endpoint:    cfg.SNS.Endpoint,
		TopicPrefix: cfg.SNS.TopicPrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SNS client: %w", err)
	}

	log.Println("Ensuring SNS topics...")
	topicARNs, err := platformsns.EnsureTopics(ctx, client, cfg.SNS.TopicPrefix, platformsns.DefaultTopicDefinitions())
	if err != nil {
		return nil, fmt.Errorf("failed to ensure SNS topics: %w", err)
	}

	log.Println("Ensuring SNS subscriptions...")
	subs := buildSubscriptions(sqsResources.QueueURLs)
	if err := platformsns.EnsureSubscriptions(ctx, client, sqsResources.Client, topicARNs, subs); err != nil {
		return nil, fmt.Errorf("failed to ensure SNS subscriptions: %w", err)
	}

	producer := platformsns.NewSNSProducer(client, topicARNs)

	log.Println("SNS initialized successfully")

	return &SNSResources{
		Client:    client,
		Producer:  producer,
		TopicARNs: topicARNs,
	}, nil
}

func buildSubscriptions(queueURLs map[string]string) []platformsns.SubscriptionDefinition {
	subs := make([]platformsns.SubscriptionDefinition, 0, len(platformsqs.DefaultConsumerQueueDefinitions()))
	for _, def := range platformsqs.DefaultConsumerQueueDefinitions() {
		if def.SubscribesToEvent == "" {
			continue
		}
		queueURL, ok := queueURLs[def.ConsumerName]
		if !ok {
			continue
		}
		subs = append(subs, platformsns.SubscriptionDefinition{
			EventName:    def.SubscribesToEvent,
			ConsumerName: def.ConsumerName,
			QueueURL:     queueURL,
		})
	}
	return subs
}
