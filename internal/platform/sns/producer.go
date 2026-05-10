package sns

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/uuid"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

type SNSProducer struct {
	client    *awssns.Client
	topicARNs map[string]string
}

func NewSNSProducer(client *awssns.Client, topicARNs map[string]string) *SNSProducer {
	return &SNSProducer{
		client:    client,
		topicARNs: topicARNs,
	}
}

func (p *SNSProducer) Dispatch(ctx context.Context, evt event.Event) error {
	topicARN, ok := p.topicARNs[evt.Name]
	if !ok {
		return fmt.Errorf("no topic ARN configured for event: %s", evt.Name)
	}

	body, err := json.Marshal(evt.Payload)
	if err != nil {
		return fmt.Errorf("error serializing event payload: %w", err)
	}

	groupID := extractGroupID(evt)
	dedupID := uuid.New().String()

	_, err = p.client.Publish(ctx, &awssns.PublishInput{
		TopicArn:               aws.String(topicARN),
		Message:                aws.String(string(body)),
		MessageGroupId:         aws.String(groupID),
		MessageDeduplicationId: aws.String(dedupID),
	})
	if err != nil {
		return fmt.Errorf("sns publish failed: %w", err)
	}

	return nil
}

func (p *SNSProducer) Close() {}

func extractGroupID(evt event.Event) string {
	if payload, ok := evt.Payload.(map[string]any); ok {
		if id, ok := payload["id"].(string); ok && id != "" {
			return id
		}
	}
	return uuid.New().String()
}
