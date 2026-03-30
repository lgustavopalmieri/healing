package sqs

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type QueueDefinition struct {
	EventName         string
	Suffix            string
	DLQSuffix         string
	MaxReceiveCount   int
	VisibilityTimeout int
	RetentionPeriod   int
}

func DefaultQueueDefinitions() []QueueDefinition {
	return []QueueDefinition{
		{
			EventName:         "specialist.created",
			Suffix:            "created",
			DLQSuffix:         "created-dlq",
			MaxReceiveCount:   3,
			VisibilityTimeout: 30,
			RetentionPeriod:   1209600,
		},
		{
			EventName:         "specialist.updated",
			Suffix:            "updated",
			DLQSuffix:         "updated-dlq",
			MaxReceiveCount:   3,
			VisibilityTimeout: 30,
			RetentionPeriod:   1209600,
		},
	}
}

func queueName(prefix, suffix string) string {
	return prefix + "-" + suffix + ".fifo"
}

// EnsureQueues creates all queues idempotently and returns a map of eventName -> queueURL.
// Safe to call from multiple pods simultaneously.
func EnsureQueues(ctx context.Context, client *sqs.Client, prefix string, definitions []QueueDefinition) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	urls := make(map[string]string)

	for _, def := range definitions {
		dlqName := queueName(prefix, def.DLQSuffix)
		dlqURL, err := createFIFOQueue(ctx, client, dlqName, def.VisibilityTimeout, def.RetentionPeriod)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure DLQ %s: %w", dlqName, err)
		}
		log.Printf("DLQ ready: %s", dlqName)

		dlqArn, err := getQueueArn(ctx, client, dlqURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get DLQ ARN for %s: %w", dlqName, err)
		}

		mainName := queueName(prefix, def.Suffix)
		redrivePolicy := fmt.Sprintf(`{"deadLetterTargetArn":"%s","maxReceiveCount":"%d"}`, dlqArn, def.MaxReceiveCount)

		mainURL, err := createFIFOQueueWithRedrive(ctx, client, mainName, def.VisibilityTimeout, def.RetentionPeriod, redrivePolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure queue %s: %w", mainName, err)
		}
		log.Printf("Queue ready: %s", mainName)

		urls[def.EventName] = mainURL
	}

	return urls, nil
}

func createFIFOQueue(ctx context.Context, client *sqs.Client, name string, visibilityTimeout, retentionPeriod int) (string, error) {
	out, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(name),
		Attributes: map[string]string{
			string(types.QueueAttributeNameFifoQueue):                 "true",
			string(types.QueueAttributeNameContentBasedDeduplication): "true",
			string(types.QueueAttributeNameVisibilityTimeout):         fmt.Sprintf("%d", visibilityTimeout),
			string(types.QueueAttributeNameMessageRetentionPeriod):    fmt.Sprintf("%d", retentionPeriod),
		},
	})
	if err != nil {
		return "", err
	}
	return *out.QueueUrl, nil
}

func createFIFOQueueWithRedrive(ctx context.Context, client *sqs.Client, name string, visibilityTimeout, retentionPeriod int, redrivePolicy string) (string, error) {
	out, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(name),
		Attributes: map[string]string{
			string(types.QueueAttributeNameFifoQueue):                 "true",
			string(types.QueueAttributeNameContentBasedDeduplication): "true",
			string(types.QueueAttributeNameVisibilityTimeout):         fmt.Sprintf("%d", visibilityTimeout),
			string(types.QueueAttributeNameMessageRetentionPeriod):    fmt.Sprintf("%d", retentionPeriod),
			string(types.QueueAttributeNameRedrivePolicy):             redrivePolicy,
		},
	})
	if err != nil {
		return "", err
	}
	return *out.QueueUrl, nil
}

func getQueueArn(ctx context.Context, client *sqs.Client, queueURL string) (string, error) {
	out, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameQueueArn},
	})
	if err != nil {
		return "", err
	}

	arn, ok := out.Attributes[string(types.QueueAttributeNameQueueArn)]
	if !ok {
		return "", fmt.Errorf("QueueArn attribute not found")
	}
	return arn, nil
}

// ResolveQueueURL gets the URL for an existing queue by prefix and suffix.
func ResolveQueueURL(ctx context.Context, client *sqs.Client, prefix, suffix string) (string, error) {
	name := queueName(prefix, suffix)
	out, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	if err != nil {
		return "", fmt.Errorf("failed to resolve queue URL for %s: %w", name, err)
	}
	return *out.QueueUrl, nil
}

// EventSuffix extracts the suffix from an event name (e.g. "specialist.created" -> "created").
func EventSuffix(eventName string) string {
	parts := strings.Split(eventName, ".")
	if len(parts) < 2 {
		return eventName
	}
	return parts[len(parts)-1]
}
