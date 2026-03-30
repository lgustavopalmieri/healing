package sqs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

type SQSProducer struct {
	client    *awssqs.Client
	queueURLs map[string]string
}

func NewSQSProducer(client *awssqs.Client, queueURLs map[string]string) *SQSProducer {
	return &SQSProducer{
		client:    client,
		queueURLs: queueURLs,
	}
}

func (p *SQSProducer) Dispatch(ctx context.Context, evt event.Event) error {
	queueURL, ok := p.queueURLs[evt.Name]
	if !ok {
		return fmt.Errorf("no queue URL configured for event: %s", evt.Name)
	}

	value, err := json.Marshal(evt.Payload)
	if err != nil {
		return fmt.Errorf("error serializing event payload: %w", err)
	}

	groupID := extractGroupID(evt)
	dedupID := uuid.New().String()

	_, err = p.client.SendMessage(ctx, &awssqs.SendMessageInput{
		QueueUrl:               aws.String(queueURL),
		MessageBody:            aws.String(string(value)),
		MessageGroupId:         aws.String(groupID),
		MessageDeduplicationId: aws.String(dedupID),
	})
	if err != nil {
		return fmt.Errorf("sqs send failed: %w", err)
	}

	return nil
}

func (p *SQSProducer) Close() {}

func extractGroupID(evt event.Event) string {
	if payload, ok := evt.Payload.(map[string]any); ok {
		if id, ok := payload["id"].(string); ok && id != "" {
			return id
		}
	}
	return uuid.New().String()
}
